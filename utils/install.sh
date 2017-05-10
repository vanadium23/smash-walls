#!/bin/bash

mkdir -p ~/.config/systemd/user/
ln -sT ($PWD)/rotate_walls.sh ~/.config/systemd/user/rotate_walls.sh  # don't like it, but no sudo
ln -sT ($PWD)/gnome-background-change.service ~/.config/systemd/user/gnome-background-change.service
ln -sT ($PWD)/gnome-background-change.timer ~/.config/systemd/user/gnome-background-change.timer

systemctl --user enable gnome-background-change.timer
systemctl --user start gnome-background-change.timer

systemctl --user list-timers
