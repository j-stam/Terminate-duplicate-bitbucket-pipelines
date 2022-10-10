# Terminate duplicate bitbucket pipelines
This script terminates older bitbucket pipelines to the same branch which are already running or queued.

## Required env variables

### Bitbucket default
https://support.atlassian.com/bitbucket-cloud/docs/variables-and-secrets/
```
BITBUCKET_WORKSPACE
BITBUCKET_REPO_SLUG
BITBUCKET_BRANCH
BITBUCKET_PIPELINE_UUID
BITBUCKET_BUILD_NUMBER
```

### Custom
```
TDP_BITBUCKET_BASIC_AUTH="base64 encoded username:app_key"
```