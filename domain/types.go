package domain

type Response struct {
	Hosts          []host `json:"host_list"`
	Total_hosts    int    `json:"total_matching"`
	Total_returned int    `json:"total_returned"`
}

type host struct {
	Name           string         `json:"name"`
	Tags_by_source tags_by_source `json:"tags_by_source"`
}

type tags_by_source struct {
	Tags    []string `json:"Datadog"`
	TagsGCP []string `json:"Google Cloud Platform"`
}

type Corporate_tags struct {
	Host             string
	Env              string
	Team             string
	Sre_team         string
	Cc_name          string
	Cc_id            string
	Cc_id_1          string
	Country          string
	Bs_dom           string
	Bs_dom_1         string
	Bs_cap           string
	Bs_cap_1         string
	Ind              string
	Bu               string
	Bu_1             string
	Cluster_name     string
	Cluster_name_gcp string
	Project_name_gcp string
	Project_id_gcp   string
}
