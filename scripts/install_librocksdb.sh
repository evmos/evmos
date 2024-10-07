#!/bin/bash

ROCKSDB_VERSION=${1:-"9.5.2"}

# Check if RocksDB is already installed
if [[ $(find /usr/lib -name "librocksdb.so.${ROCKSDB_VERSION}" -print -quit) ]]; then
	read -r -p "RocksDB version ${ROCKSDB_VERSION} is already installed. Do you want to reinstall it? (yes/no): " choice
	case "$choice" in
	y | yes | Yes | YES)
		echo "Reinstalling RocksDB..."
		rm -rf /usr/lib/librocksdb*
		;;
	n | no | No | NO)
		echo "Skipping RocksDB installation."
		exit 0
		;;
	*)
		echo "Invalid choice. Please enter 'yes' or 'no'."
		exit 1
		;;
	esac
else
	# RocksDB is not installed, proceed with installation
	echo "RocksDB is not installed. Proceeding with installation..."
fi

# Check the OS type and perform different actions
if [[ $(uname) == "Linux" ]]; then
	# Check Linux distribution
	if [[ -f /etc/os-release ]]; then
		source /etc/os-release

		if [[ "$ID" == "ubuntu" ]]; then
			# Ubuntu specific dep installation
			echo "Installing RocksDB dependencies..."
			apt-get install libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev build-essential clang

		elif [[ "$ID" == "alpine" ]]; then
			# Alpine specific dep installation
			echo "Installing RocksDB dependencies..."
			# 1. Install dependencies
			echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories
			apk add --update --no-cache cmake bash perl g++
			apk add --update --no-cache zlib zlib-dev bzip2 bzip2-dev snappy snappy-dev lz4 lz4-dev zstd@testing zstd-dev@testing libtbb-dev@testing libtbb@testing
			# 2. Install latest gflags
			cd /tmp &&
				git clone https://github.com/gflags/gflags.git &&
				cd gflags &&
				mkdir build &&
				cd build &&
				cmake -DBUILD_SHARED_LIBS=1 -DGFLAGS_INSTALL_SHARED_LIBS=1 .. &&
				make install &&
				rm -rf /tmp/gflags
		else
			echo "Linux distribution not supported"
			exit 1
		fi

		# 3. Install Rocksdb (same for any linux distribution)
		cd /tmp &&
			git clone -b v"${ROCKSDB_VERSION}" --single-branch https://github.com/facebook/rocksdb.git &&
			cd rocksdb &&
			PORTABLE=1 WITH_JNI=0 WITH_BENCHMARK_TOOLS=0 WITH_TESTS=1 WITH_TOOLS=0 WITH_CORE_TOOLS=1 WITH_BZ2=1 WITH_LZ4=1 WITH_SNAPPY=1 WITH_ZLIB=1 WITH_ZSTD=1 WITH_GFLAGS=0 USE_RTTI=1 \
				make shared_lib &&
			cp librocksdb.so* /usr/lib/ &&
			cp -r include/* /usr/include/ &&
			rm -rf /tmp/rocksdb
	else
		echo "Cannot determine Linux distribution."
		exit 1
	fi

elif [[ $(uname) == "Darwin" ]]; then
	# macOS-specific actions
	xcode-select --install
	brew tap homebrew/versions
	brew install gcc7 --use-llvm
	brew install rocksdb
else
	echo "Unsupported OS."
	exit 1
fi
