# Set Variables

From a terminal session inside your Cloud9 environment, prepare your EC2 variables file:
```bash
cat > ~/.env << EOF
alias k="kubectl"                                                           # a common shortcut for the CLI

export AWS_ACCOUNT_ID=$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document|grep accountId|awk -F\" '{print $4}')
export AWS_DEFAULT_REGION=$(curl --silent http://169.254.169.254/latest/meta-data/placement/region)
export AWS_PAGER=                                                           # intentionally blank

export EKS_GITHUB_PUBLIC_REPO=https://github.com/amcginlay/eks-demos.git    # if you fork this repo, change this!
export EKS_CLUSTER_NAME=dev
export EKS_K8S_VERSION=1.20

export EKS_APP=php-echo
export EKS_APP_NS=\${EKS_APP}
export EKS_APP_VERSION=1.0.42
export EKS_APP_VERSION_NEXT=1.0.43
export EKS_APP_ECR_REPO=\${AWS_ACCOUNT_ID}.dkr.ecr.\${AWS_DEFAULT_REGION}.amazonaws.com/\${EKS_APP}
EOF
```

Ensure these variables get set into every bash session then set the variables into your current shell so we can use them immediately:
```bash
echo "source ~/.env" >> ~/.bashrc
source ~/.env
```

Familiarize yourself with these variable settings
```bash
env | sort | grep "AWS\|EKS"
```

[Return To Main Menu](/README.md)
