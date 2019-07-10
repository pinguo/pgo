#!/usr/bin/env sh
###########################################
#  go test                                #
#  for Windows/Mac/Linux                  #
#                                         #
#  Version: 1.0                           #
#  Author: yangbing@camera360.com         #
#  Date: 2019/07/05                       #
###########################################

# 要go test 包名字，如果是mod模式下  module_name/package_name,可修改
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
exec_mod="default"

if [[ $# > 0 ]];then
    go_app_path=$1
fi

if [[ $# > 1 ]];then
    exec_mod="custom"
fi

# PgoTestAppBasePath提供给pgo框架使用，必须设置，conf配置文件夹的上一级目录
export PgoTestAppBasePath=$go_app_path


exec_default () {
    test_path_dir=$go_app_path
    cover_dir=${test_path_dir}/coverage

    if [[ ! -d $cover_dir ]];then
        mkdir $cover_dir
    fi

    if [[ ! -d ${go_app_path}/src ]];then
        test_path_dir=${go_app_path}/src
    fi

    pkg_str=""
    cover_pkg=""
    for test_path_name in ${arr_test_path_dir_name[@]}
    do
        pkg_str=${pkg_str}" "${test_path_name}"/..."
        if [[ ${cover_pkg} == "" ]];then
        cover_pkg=${test_path_name}"/..."
        else
        cover_pkg=${cover_pkg}","${test_path_name}"/..."
        fi

    done

    go test -coverprofile=${cover_dir}/coverage.data -coverpkg=$cover_pkg $pkg_str
    go tool cover -html=${cover_dir}/coverage.data -o ${cover_dir}/coverage.html
    go tool cover -func=${cover_dir}/coverage.data -o ${cover_dir}/coverage.txt
}

case "$exec_mod" in
  default)
    exec_default
    ;;
  custom)
    ${@:2}
    ;;
  
esac
