# Variables

Prepare you environment file.

```bash
cat > ~/.env << EOF
export AWS_DEFAULT_REGION=$(curl --silent http://169.254.169.254/latest/meta-data/placement/region)
export AWS_PAGER=""
cluster=dev
k8s_version=1.18
namespace=demo
app_name=demo
app_version=1.0.42
EOF
```

Ensure these variables get set into every bash session.

```bash
echo "source ~/.env" >> ~/.bashrc
```

Load the variables into your current shell so we can use them immediately.

```bash
source ~/.env
```

[Return To Main Menu](../README.md)
