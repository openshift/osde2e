# Scripts

Various scripts that support the execution and day-to-day operation of osde2e.

## metrics-sync.sh

This script syncs the prometheus export files located in the osde2e-metrics S3 bucket with the Datahub prometheus instance via Datahub prometheus pushgateway.

### Prerequisites

The following environment variables are expected to be set:

| Environment variable     | Description                                        |
|--------------------------|----------------------------------------------------|
| PUSHGATEWAY\_URL         | The URL to use for pushing metrics.                |

Python 2+ and virtualenv are necessary to install the latest awscli. It's very possible that disparate versions of Python will work -- as long as the aws CLI runs, this script should run.

#### Note about AWS configuration

The AWS CLI must be configured for s3 interaction. This can be done a number of ways, the most straightforward of which being using environment variables:

| Environment variable     | Description                                        |
|--------------------------|----------------------------------------------------|
| AWS\_ACCESS\_KEY\_ID     | The AWS access key ID for interacting with S3.     |
| AWS\_SECRET\_ACCESS\_KEY | The AWS secret access key for interacting with S3. |
| AWS\_REGION              | The AWS region.                                    |

Most likely this script will be running with the environment variables set. However, this is not the only way: you can use 
[config and credentials files](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html),
IAM instance profiles (if running on an AWS instance), or a valid combination. See 
[AWS's documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) for more information.
