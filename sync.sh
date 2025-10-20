#!/bin/bash

REPO_URL="https://github.com/hackclub/milkyway.git"

rm -rf ressources/synced/*

# Initialize repo
git init tmp_milkyway
cd tmp_milkyway || exit

git remote add origin "$REPO_URL"
git config core.sparseCheckout true

# Sparse checkout setup
echo "static/room/" > .git/info/sparse-checkout
echo "static/projects/" >> .git/info/sparse-checkout

# Fetch and checkout
git fetch --depth=10 origin main
git checkout main

cd ..

mkdir -p ressources/synced/
mkdir -p ressources/synced/room/
mkdir -p ressources/synced/projects/
mv tmp_milkyway/static/room/* ressources/synced/room/
mv tmp_milkyway/static/projects/* ressources/synced/projects/
mv ressources/synced/room/floor ressources/synced/
rm -rf tmp_milkyway
