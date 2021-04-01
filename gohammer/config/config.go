package config

type Config struct {
	TestProfiles []TestProfile `json:"profiles"`

	// Indicates that test profiles will be run concurrently
	// if it is true.
	//
	// Do Not Mix With TestProfile Concurrent Option.
	// This concurrent option is related with all test profiles
	// (Test profiles are going to be run immediately or have to
	// wait other test profiles to be finished?)
	Concurrent bool `json:"concurrent`
}

type TestProfile struct {
	Name  string       `json:"name"`
	Nodes []NodeConfig `json:"nodes"`

	// Indicates that given test profile contract will be deployed
	// concurrently if its value is true.
	//
	// Do Not Mix With Overall Concurrent Option.
	// This concurrent option is related with just the given test profile,
	// not all test profile.
	Concurrent bool `json:"concurrent`

	//TODO: change key
	RoundRobin bool `json:"roundRobin"`
}

type NodeConfig struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Cipher         string `json:"cipher"`
	DeployCounts   []int  `json:"deployCounts"`
	DeployInterval string `json:"deployInterval"`
}
