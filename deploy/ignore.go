package deploy

import "github.com/launchpad-project/cli/configstore"

func getIgnoreList(cs *configstore.Store) []string {
	var ignoreList, err = cs.GetInterface("deploy_ignore")

	var ignorePatterns []string

	switch err {
	case configstore.ErrConfigKeyNotFound:
		ignorePatterns = []string{}
	case nil:
		var ii = ignoreList.([]interface{})
		var b = make([]string, len(ii))

		for i := range ii {
			b[i] = ii[i].(string)
		}

		ignorePatterns = b
	default:
		panic(err)
	}

	return ignorePatterns
}
