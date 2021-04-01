<?php
  \$x = 0.0001; for (\$i = 0; \$i <= 1000000; \$i++) { \$x += sqrt(\$x); } // simulate some work
  \$ec2_instance = shell_exec('curl http://169.254.169.254/latest/meta-data/instance-id');
  \$ec2_ip = shell_exec('curl http://169.254.169.254/latest/meta-data/local-ipv4');
  \$localhost_ip = shell_exec('hostname -i | tr -d \'\n\'');
  echo '{ "date": "' . date("c") . '", "version": "' . getenv("VERSION") . '", "ec2Instance": "' . \$ec2_instance . '", "ec2IP": "' . \$ec2_ip . '", "localhostIP": "' . \$localhost_ip . '" }' . "\n";
?>
