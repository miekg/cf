all:
	( cd cmd/cffmt; go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" )
	( cd cmd/cfgroup; go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" )

install:
	( cd cmd/cffmt; go install )
	( cd cmd/cfgroup; go install )

man:
	( cd cmd/cffmt; mmark -man cffmt.1.md > cffmt.1 )
	( cd cmd/cfgroup; mmark -man cfgroup.1.md > cfgroup.1 )
