package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/hekmon/cunits/v2"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

var (
	version        = "dev"
	toRecover      cunits.Bits
	spaceRecovered cunits.Bits = 0
	toDrop         []int64
)

// Get domain name from a full url
func url2domain(fullurl string) string {
	url, err := url.Parse(fullurl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	parts := strings.Split(url.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	return domain
}

func main() {

	viper.SetDefault("transmission.host", "127.0.0.1")
	viper.SetDefault("transmission.port", "9091")
	viper.SetDefault("transmission.user", "username")
	viper.SetDefault("transmission.pass", "password")
	viper.SetDefault("path", "/")
	viper.SetDefault("free_giga", 250)
	viper.SetDefault("tracker_keep", 2)

	pflag.StringP("config", "c", "", "Config file path")
	pflag.Bool("cron", false, "Same as --do flag but no ouput when enough free space")
	pflag.Bool("do", false, "Commit torrent deletion")
	pflag.IntP("free_giga", "f", 250, "Target GiB free")
	pflag.IntP("tracker_keep", "k", 2, "Torrent to keep per tracker")
	pflag.BoolP("help", "h", false, "This help message")
	pflag.BoolP("version", "v", false, "This help message")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.GetBool("help") {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(1)
	}

	if viper.GetBool("version") {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	// Load config from file
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		viper.AddConfigPath("$HOME/.config")
		viper.AddConfigPath(".")
		viper.SetConfigName("btcleaner") // Register config file name (no extension)
		viper.SetConfigType("yaml")      // Look for specific type
	}
	viper.ReadInConfig()

	conf_host := viper.GetString("transmission.host")
	conf_port := viper.GetUint16("transmission.port")
	conf_user := viper.GetString("transmission.user")
	conf_pass := viper.GetString("transmission.pass")
	conf_path := viper.GetString("path")
	conf_freegiga := viper.GetInt("free_giga")
	conf_trackerkeep := viper.GetInt("tracker_keep")
	conf_exclude := viper.GetStringSlice("exclude")
	conf_cron := viper.GetBool("cron")
	conf_do := viper.GetBool("do")

	if conf_cron && conf_do {
		fmt.Println("ERROR: Flags --cron and --do are mutually exclusive because --cron is a less verbose version of --do")
		os.Exit(1)
	}

	// Get wanted space in cunits
	targetSpace, _ := cunits.Parse(fmt.Sprint(conf_freegiga) + " GiB")

	// Instanciate Transmission connection
	transmissionbt, _ := transmissionrpc.New(conf_host, conf_user, conf_pass,
		&transmissionrpc.AdvancedConfig{
			Port: conf_port,
		})

	// Get Free space
	freeSpace, err := transmissionbt.FreeSpace(context.TODO(), conf_path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Choose if we need to recover space
	if freeSpace >= targetSpace {
		if !conf_cron {
			fmt.Println("We have ", freeSpace, "Free. It's above ", targetSpace, ". No need to free space !")
		}
		os.Exit(0)
	} else {
		toRecover = targetSpace - freeSpace
		fmt.Println(freeSpace, "left, target is", targetSpace, "free. We need to recover ", toRecover, ".")
		fmt.Println("---")
	}

	// Fetch all torrent
	torrents, err := transmissionbt.TorrentGetAll(context.TODO())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// sort by date Added
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].AddedDate.Before(*torrents[j].AddedDate)
	})

	// Count torrent per tracker
	torrentPerTracker := make(map[string]int)
	for _, torrent := range torrents {
		torrentPerTracker[url2domain(torrent.Trackers[0].Announce)]++
	}

	// iterate to select torrent to drop
	for _, torrent := range torrents {
		thisTracker := url2domain(torrent.Trackers[0].Announce)
		if slices.Contains(conf_exclude, *torrent.Name) {
			fmt.Println("Excluded because on exclude list:", *torrent.Name)
		} else if torrentPerTracker[thisTracker] > conf_trackerkeep {
			toDrop = append(toDrop, *torrent.ID)
			torrentPerTracker[thisTracker]--
			spaceRecovered = spaceRecovered + *torrent.SizeWhenDone
			fmt.Println("Recovered", spaceRecovered, "/", toRecover, ": (", *torrent.SizeWhenDone, ") [", thisTracker, "]", *torrent.Name)
			if spaceRecovered >= toRecover {
				break
			}
		} else {
			fmt.Println("Excluded because only", torrentPerTracker[thisTracker], "torrents left on ", thisTracker, ":", *torrent.Name)
		}
	}

	if conf_do || conf_cron {
		body := transmissionrpc.TorrentRemovePayload{IDs: toDrop, DeleteLocalData: true}
		err := transmissionbt.TorrentRemove(context.TODO(), body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		freeSpace, err := transmissionbt.FreeSpace(context.TODO(), conf_path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println(len(toDrop), "torrent deleted. ", freeSpace, "Free.")
	} else {
		fmt.Println("---")
		fmt.Println("DRY RUN ! Please use --do flag to really drop torrents.")
	}

}
