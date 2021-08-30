#!/bin/bash

podman build . -t smb:test-03
podman run -it -h ad --name smb --net=host smb:test-03 -n smb
