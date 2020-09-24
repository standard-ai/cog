#!/bin/bash

set -x

# Run cog to SSH to cog_target
../../build/cog -v ssh -u ubuntu cog_target
