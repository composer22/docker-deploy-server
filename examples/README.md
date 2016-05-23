### Additional Configuration Examples

* docker-deploy-server.service - example systemd service file
* docker-deploy-server.yml - example configuration file template

#### docker-deploy-server.service

Install instructions.

1) Place the config wherever you have this server code residing.

2) Create a symlink to it in /etc/systemd/system/multi-user.target.wants

ex:
sudo ln -s /home/ubuntu/bin/docker-deploy-server/docker-deploy-server.service \
   docker-deploy-server.service

3) systemctl daemon-reload

4) systemctl start docker-deploy-server.service


