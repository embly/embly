#!/bin/sh

cp /opt/* /run
cd /run
ls -lah 
terraform "$@"