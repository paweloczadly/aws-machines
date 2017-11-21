# aws-machines

Lists all EC2 instances (name, instance type, AZ) from all regions per AWS accounts and saves it to CSV file.

## Usage

Go to releases. Download proper binary for your OS. For Linux run the following:

```
$ mv aws-machines-linux-amd64 aws-machines
$ chmod +x aws-machines
$ ./aws-mchines accessKeys.csv output.csv
```

accessKeys.csv should look as follow:

```
<account-id-1>,<access-key-id>,<secret-access-key>
<account-id-2>,<access-key-id>,<secret-access-key>
...
```

## Local development

Set up aws-sdk-go:

```
$ go get -u github.com/aws/aws-sdk-go/...
```

Run the project locally:

```
$ go run aws-machines.go accessKey.csv output.csv
```

Build the project locally:

```
$ go build
```
