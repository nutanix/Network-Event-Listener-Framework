package config

type F5Config struct {
	F5InstanceConfig struct {
		IP       string `json:"ip"`
		Password string `json:"password"`
		Pools    []struct {
			PoolMembers []string `json:"pool_members"`
			PoolName    string   `json:"pool_name"`
		} `json:"pools"`
		Port        string `json:"port"`
		Serviceport string `json:"serviceport"`
		Username    string `json:"username"`
	} `json:"f5_instance_config"`
	NutanixClusterConfig struct {
		IP       string `json:"ip"`
		Password string `json:"password"`
		Port     string `json:"port"`
		Username string `json:"username"`
	} `json:"nutanix_cluster_config"`
}
