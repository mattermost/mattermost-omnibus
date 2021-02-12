module omnitests

go 1.14

require (
	github.com/magefile/mage v1.10.0
	github.com/mattermost/mattermost-omnibus/mmomni v0.0.0
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20201023074133-8c63eb7232d6
	github.com/stretchr/testify v1.6.1
)

replace github.com/mattermost/mattermost-omnibus/mmomni => ../mmomni
