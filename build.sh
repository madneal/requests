# build for linux
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
#!/bin/bash

function obtain_git_branch {
  br=`git branch | grep "*"`
  echo ${br/* /}
}
result=`obtain_git_branch`
echo Current git branch is $result
if [ "$result" == "master" ]
 then
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
fi

git checkout cel
result=`obtain_git_branch`
echo Current git branch is $result
if [ "$result" == "cel" ]
  then
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cel
fi

