# Prepare your CloudShell environment

This environment will be created in Oregon (us-west-2) simply because it's tried and tested there. Feel free to modify as appropriate. 

Navigate to https://us-west-2.console.aws.amazon.com/cloudshell

Update the AWS CLI.
```bash
rm -rf /usr/local/bin/aws 2> /dev/null
rm -rf /usr/local/aws-cli 2> /dev/null
rm -rf aws awscliv2.zip 2> /dev/null
curl --silent "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install --update
```

[Return To Main Menu](/README.md)
