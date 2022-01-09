# Set Variables

From a terminal session inside your Cloud9 environment, prepare your EC2 variables file.
```bash
cat > ~/.env << EOF
alias k="kubectl"                                                           # a common shortcut for the CLI

export AWS_ACCOUNT_ID=$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document|grep accountId|awk -F\" '{print $4}')
export AWS_DEFAULT_REGION=$(curl --silent http://169.254.169.254/latest/meta-data/placement/region)
export AWS_PAGER=                                                           # intentionally blank

export EKS_GITHUB_USER=amcginlay                                            # if you fork this repo, change this!
export EKS_CLUSTER_NAME=dev
export EKS_K8S_VERSION=1.20

export EKS_ECR_REGISTRY=\${AWS_ACCOUNT_ID}.dkr.ecr.\${AWS_DEFAULT_REGION}.amazonaws.com
EOF
```

Ensure these variables get set into every bash session then set the variables into your current shell so you can use them immediately.
```bash
echo "source ~/.env" >> ~/.bashrc
source ~/.env
```

Familiarize yourself with these variable settings
```bash
env | sort | grep "AWS\|EKS"
```

[Return To Main Menu](/README.md)
