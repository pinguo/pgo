#!/usr/bin/env sh
###########################################
#  go test                                #
#  for Windows/Mac/Linux                          #
#                                         #
#  Version: 1.0                           #
#  Author: yangbing@camera360.com         #
#  Date: 2019/07/05                       #
###########################################

# 要go test目录名字,可修改
arr_test_path_dir_name=(
    Command
    Controller
    Lib
    Model
    Service
    Struct
    Test
)

go_app_path=`pwd`
# PgoTestAppBasePath提供给pgo框架使用，必须设置，conf配置文件夹的上一级目录
export PgoTestAppBasePath=$go_app_path

exec_mod="default"

if [[ $# > 0 ]];then
    exec_mod="custom"
fi


exec_default () {
    test_path_dir=$go_app_path
    cover_dir=${test_path_dir}/coverage

    if [ ! -d $cover_dir ]
    then
        mkdir $cover_dir
    fi

    if [ ! -d ${go_app_path}/src ]
    then
        test_path_dir=${go_app_path}/src
    fi

    pkg_str=""
    cover_pkg=""
    for test_path_name in ${arr_test_path_dir_name[@]}
    do
        pkg_str=${pkg_str}" "${test_path_name}"/..."
        cover_pkg=${cover_pkg}","${test_path_name}"/..."
    done

    echo

    go test -coverprofile=${cover_dir}/coverage.data -coverpkg=$cover_pkg $pkg_str
    go tool cover -html=${cover_dir}/coverage.data -o ${cover_dir}/coverage.html
    go tool cover -func=${cover_dir}/coverage.data -o ${cover_dir}/coverage.txt
}


case "$exec_mod" in
  default)
    exec_default
    ;;
  custom)
    $@
    ;;
  
esac
