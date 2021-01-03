#!/bin/sh

cd $(dirname "$0")
gtts-cli -l zh-cn "$1" > speak.mp3
play speak.mp3 # or cvlc --play-and-exit speak.mp3 
rm speak.mp3
