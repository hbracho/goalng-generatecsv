package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	domains "hbracho/datadog/generate-csv/domain"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	godotenv.Load(".env")

}
func buildFile(data []domains.Corporate_tags) {
	csvFile, err := os.Create("resources/host_dd.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	header := []string{"host", "env", "team", "sre_team", "cc-name", "cc-id", "country", "bs-dom", "bs-cap", "ind", "bu", "cluster-name", "project_id_gcp", "project_name_gcp"}

	_ = csvwriter.Write(header)
	for _, row := range data {
		var raw []string
		raw = append(raw, row.Host)
		raw = append(raw, row.Env)
		raw = append(raw, row.Team)
		raw = append(raw, row.Sre_team)
		raw = append(raw, row.Cc_name)
		raw = append(raw, getValue(row.Cc_id, row.Cc_id_1))
		raw = append(raw, row.Country)
		raw = append(raw, getValue(row.Bs_dom, row.Bs_dom_1))
		raw = append(raw, getValue(row.Bs_cap, row.Bs_cap_1))
		raw = append(raw, row.Ind)
		raw = append(raw, getValue(row.Bu, row.Bu_1))
		raw = append(raw, getValue(row.Cluster_name, row.Cluster_name_gcp))
		raw = append(raw, row.Project_id_gcp)
		raw = append(raw, row.Project_name_gcp)

		_ = csvwriter.Write(raw)
	}
	csvwriter.Flush()
}

func getValue(v1 string, v2 string) string {
	if len(v1) > 0 {
		return v1
	} else {
		return v2
	}
}

func invokeDataDog(from int, limit int) domains.Response {
	log.WithFields(
		log.Fields{"start": from,
			"limit": limit},
	).Debug("invoking datadog api")

	url := fmt.Sprintf(os.Getenv("URL_PATH"), from, limit)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("DD-API-KEY", os.Getenv("DD_API_KEY"))
	req.Header.Add("DD-APPLICATION-KEY", os.Getenv("DD_APPLICATION_KEY"))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject domains.Response
	json.Unmarshal(responseData, &responseObject)

	log.Debug("finished call datadog api")

	return responseObject
}

func buildCSV() {
	log.Debug("staring the project")
	start := 0
	limit := 1000
	responsedata := invokeDataDog(start, limit)
	total_hosts := responsedata.Total_hosts
	total_returned := responsedata.Total_returned

	raw_data := buildraw(responsedata)

	log.WithFields(log.Fields{"start": start, "total_hosts": total_hosts, "total_returned": total_returned}).Info("invoking api from main")

	for total_hosts > total_returned {

		start = start + limit + 1
		a := invokeDataDog(start, limit)
		log.WithFields(log.Fields{"start": start, "total_hosts": a.Total_hosts, "total_returned": a.Total_returned}).Info("invoking api from main")
		rd := buildraw(a)
		raw_data = append(raw_data, rd...)
		total_hosts = a.Total_hosts
		total_returned = total_returned + a.Total_returned + 1
	}
	buildFile(raw_data)
	log.Info("Finished process")
}

func buildraw(data domains.Response) []domains.Corporate_tags {
	log.Debug("building data corporate")
	var result []domains.Corporate_tags

	for i := 0; i < len(data.Hosts); i++ {
		h := data.Hosts[i]

		m := make(map[string]string)
		mgcp := make(map[string]string)

		for t := 0; t < len(h.Tags_by_source.Tags); t++ {
			tag_key_value := strings.Split(h.Tags_by_source.Tags[t], ":")
			if len(tag_key_value) > 1 {

				m[tag_key_value[0]] = tag_key_value[1]
			}
		}

		//log.Debug("all values of the m with its values and keys are: ", m)

		for t := 0; t < len(h.Tags_by_source.TagsGCP); t++ {

			tag_key_value_gcp := strings.Split(h.Tags_by_source.TagsGCP[t], ":")
			//fmt.Printf("hostname is %s", h.Name)
			//fmt.Printf(" value tags %s", h.Tags_by_source.TagsGCP[t])
			if len(tag_key_value_gcp) > 1 {
				mgcp[tag_key_value_gcp[0]] = tag_key_value_gcp[1]
			}
		}

		if len(mgcp) > 0 {
			log.Debug("all values of the mgcp with its values and keys are: ", mgcp)
		}

		c := domains.Corporate_tags{
			Host:             h.Name,
			Env:              m["env"],
			Team:             m["team"],
			Sre_team:         m["sre_team"],
			Cc_name:          m["cc-name"],
			Cc_id:            m["cc-id"],
			Cc_id_1:          m["cc_id"],
			Country:          m["country"],
			Bs_dom:           m["bs-dom"],
			Bs_dom_1:         m["domain"],
			Bs_cap:           m["bs-cap"],
			Bs_cap_1:         m["capability"],
			Ind:              m["ind"],
			Bu:               m["bu"],
			Bu_1:             m["business_unit"],
			Cluster_name:     m["cluster-name"],
			Cluster_name_gcp: mgcp["cluster-name"],
			Project_name_gcp: mgcp["project"],
			Project_id_gcp:   mgcp["numeric_project_id"],
		}

		result = append(result, c)
	}
	log.Debug("finished  data corporate")
	return result
}

func main() {
	buildCSV()
}
