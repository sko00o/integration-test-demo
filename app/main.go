package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"example.com/app/storage"
)

type Config struct {
	Listen  string         `mapstructure:"listen"`
	Storage storage.Config `mapstructure:"storage"`
}

type LogInfo struct {
	Timestamp int64  `json:"timestamp"`
	Host      string `json:"host"`
	Flux      int64  `json:"flux"`
}

func RunServer(cfg Config) error {
	st, err := storage.New(cfg.Storage)
	if err != nil {
		return fmt.Errorf("new storage: %w", err)
	}
	defer func() { _ = st.Close() }()

	http.HandleFunc("/incr", func(w http.ResponseWriter, r *http.Request) {
		var rd io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			rd, err = gzip.NewReader(rd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		var infos []storage.Info
		sc := bufio.NewScanner(rd)
		for sc.Scan() {
			var i LogInfo
			if err := json.Unmarshal(sc.Bytes(), &i); err == nil {
				tm := time.Unix(i.Timestamp, 0).
					Truncate(5 * time.Minute).
					UTC().
					Format("200601021504")
				infos = append(infos, storage.Info{Key: tm, Field: i.Host, Val: i.Flux})
			}
		}
		if err := sc.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := st.Incr(r.Context(), infos...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("time")
		if key == "" {
			http.Error(w, "time required", http.StatusBadRequest)
			return
		}
		field := r.URL.Query().Get("host")
		if field == "" {
			res, err := st.Client.HGetAll(r.Context(), key).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			list := make([]string, 0, len(res))
			for k, v := range res {
				list = append(list, fmt.Sprintf("%s:%s\n", k, v))
			}
			sort.Strings(list)
			for _, out := range list {
				_, _ = fmt.Fprint(w, out)
			}
		} else {
			res, err := st.Client.HGet(r.Context(), key, field).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			out := fmt.Sprintf("%s:%s\n", field, res)
			_, _ = fmt.Fprint(w, out)
		}
	})
	return http.ListenAndServe(cfg.Listen, nil)
}

func main() {
	cmd := &cobra.Command{
		Use: "app",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			viper.SetConfigFile(cfgPath)
			if err := viper.ReadInConfig(); err != nil {
				return err
			}

			var cfg Config
			if err := viper.Unmarshal(&cfg); err != nil {
				return err
			}
			return RunServer(cfg)
		},
	}
	cmd.Flags().StringP("config", "c", "", "config file path")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
