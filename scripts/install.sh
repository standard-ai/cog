#!/usr/bin/env bash

SOURCE=build/cog

if [ -f $SOURCE ] ; then
	if [ -z "$PREFIX" ] ; then
		if [ -d ~/bin ] ; then
			echo cp $SOURCE ~/bin/cog
			cp $SOURCE ~/bin/cog
		elif [ -d /usr/local/bin ] ; then
			echo cp $SOURCE /usr/local/bin/cog
			cp $SOURCE /usr/local/bin/cog
		else
			echo ~/bin and /usr/local/bin do not exist. Aborting.
			exit 1
		fi
	else
		echo cp $SOURCE $PREFIX/cog
		cp $SOURCE $PREFIX/cog
	fi
fi
