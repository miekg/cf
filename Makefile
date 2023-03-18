all:
	( cd cmd/cffmt; go build )

install:
	( cd cmd/cffmt; go install )

man:
	( cd cmd/cffmt; mmark -man cffmt.1.md > cffmt.1 )
