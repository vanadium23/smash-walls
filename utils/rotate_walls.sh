#!/bin/bash

USER=$(whoami)
PID=$(pgrep -u $USER gnome-session)
export DBUS_SESSION_BUS_ADDRESS=$(grep -z DBUS_SESSION_BUS_ADDRESS /proc/$PID/environ|cut -d= -f2-)

walls_dir=$1
if [ -z "$walls_dir" ]
    then
        year=`date +'%Y'`
        month=`date +'%m'`
        walls_dir=$HOME/Pictures/Smashing-Wallpapers/$year/$month/
fi
selection=$(find $walls_dir -type f -name "*.jpg" -o -name "*.png" | shuf -n1)
gsettings set org.gnome.desktop.background picture-uri "file://$selection"
