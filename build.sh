#! /bin/bash
function obtain_git_branch {
  br=`git branch | grep "*"`
  echo ${br/* /}
}

result=`obtain_git_branch`
echo Current git branch is $result
if [ "$result" == "master" ]
 then
   echo Building for the branch $result
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
   echo Finished building process
   git checkout cel
   echo Current git branch is `obtain_git_branch`
   echo Building for the branch `obtain_git_branch`
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cel
   echo Finished building process
   git checkout master
elif [ "$result" == "cel" ]
  then
   echo Building for the branch $result
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cel
   echo Finished building process
   git checkout master
   echo Current git branch is `obtain_git_branch`
   echo Building for the branch `obtain_git_branch`
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
   echo Finished building process
fi