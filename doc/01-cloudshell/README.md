# Prepare your CloudShell environment

This EKS cluster will be created in Oregon (us-west-2) simply because it's tried and tested there. Feel free to modify as appropriate. 

Navigate to https://us-west-2.console.aws.amazon.com/cloudshell

Cloudshell regularly tracks behind the latest AWS CLI so it will need updating whenever CloudShell is started:
```bash
rm -rf /usr/local/bin/aws 2> /dev/null
rm -rf /usr/local/aws-cli 2> /dev/null
rm -rf aws awscliv2.zip 2> /dev/null
curl --silent "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install --update
```

Failure to update the CLI can cause some later commands to fail (e.g. `aws cloud9 create-environment-ec2`)

[Return To Main Menu](/README.md)
