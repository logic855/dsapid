package sync

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/MerlinDMC/dsapid"
	"github.com/MerlinDMC/dsapid/converter"
	"github.com/MerlinDMC/dsapid/converter/dsapi"
	"github.com/MerlinDMC/dsapid/storage"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

type dsapiSyncer struct {
	source *dsapid.SyncSourceResource
	queue  chan *syncerDownloadJob

	client *http.Client
	base   *url.URL

	decoder converter.ManifestDecoder

	users     storage.UserStorage
	manifests storage.ManifestStorage
}

func (me *dsapiSyncer) Init(queue chan *syncerDownloadJob) error {
	me.queue = queue
	me.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	if v, err := url.Parse(me.source.Source); err != nil {
		return err
	} else {
		me.base = v
	}

	if v, err := dsapi.NewDecoder(me.source.Provider, me.users); err != nil {
		return err
	} else {
		me.decoder = v
	}

	log.WithFields(log.Fields{
		"name": me.source.Name,
	}).Info("initialized syncer")

	return nil
}

func (me *dsapiSyncer) Run(stop chan struct{}) error {
	delay_string := me.source.Delay
	if delay_string == "" {
		delay_string = dsapid.DefaultSyncDelay
	}

	delay, err := time.ParseDuration(delay_string)
	if err != nil {
		delay = time.Duration(time.Hour * 24)
	}

	var tick <-chan time.Time = time.After(0)

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-tick:
				log.WithFields(log.Fields{
					"name": me.source.Name,
				}).Info("sync started")

				if res, err := me.client.Get(me.source.Source); err == nil {
					var entries []dsapid.Table

					if err = json.NewDecoder(res.Body).Decode(&entries); err != nil {
						log.WithFields(log.Fields{
							"name": me.source.Name,
						}).Errorf("sync error: %s", err)
					}

				nextItem:
					for _, item := range entries {
						if manifest := me.decoder.Decode(item); manifest == nil {
							log.WithFields(log.Fields{
								"name": me.source.Name,
							}).Error("sync error: can't decode manifest")

							continue nextItem
						} else {
							job := syncerDownloadJob{
								manifest: manifest,
								files:    make([]*url.URL, 0),
							}

							manifest.SyncInfo["time"] = time.Now().Format(time.RFC3339)
							manifest.SyncInfo["from"] = me.source.Source
							manifest.SyncInfo["type"] = me.source.Type

							for _, file := range manifest.Files {
								if u, err := url.Parse(fmt.Sprintf("/datasets/%s/%s", manifest.Uuid, file.Path)); err == nil {
									job.files = append(job.files, me.base.ResolveReference(u))
								}
							}

							me.queue <- &job
						}
					}
				} else {
					log.WithFields(log.Fields{
						"name": me.source.Name,
					}).Errorf("sync error: %s", err)
				}

				log.WithFields(log.Fields{
					"name": me.source.Name,
				}).Info("sync finished")

				tick = time.After(delay)
			}
		}
	}()

	return nil
}
