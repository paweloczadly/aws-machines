package main

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	if len(os.Args) <= 2 {
		log.Fatal("Usage: ./aws-machines credentials.csv output.csv")
	}

	// Open input file:
	inputFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to open %s. Error:\n %v", os.Args[1], err)
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	// Create output file:
	outputFile, err := os.Create(os.Args[2])
	defer outputFile.Close()
	if err != nil {
		log.Fatalf("Failed to create %s. Error:\n %v", os.Args[2], err)
	}

	// Write column names to file:
	writeToFile(*outputFile, []string{"Account", "Name", "Instance Type", "Availability Zone"})

	for scanner.Scan() {
		// Setup ACCESS KEY ID and SECRET ACCESS KEY, then connect to AWS:
		inputLine := strings.Split(scanner.Text(), ",")
		accountId, accessKey, secretAccessKey := inputLine[0], inputLine[1], inputLine[2]
		connect(accountId, accessKey, secretAccessKey, *outputFile)
	}
}

func connect(accountId string, accessKey string, secretAccessKey string, outputFile os.File) {
	var wg sync.WaitGroup

	// Create AWS session and fetch available regions:
	sess, _ := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
		Region:      aws.String("eu-central-1"),
	})
	regionsApiCall := ec2.New(sess)
	regions, err := regionsApiCall.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		log.Fatalf("Failed to call DescribeRegions. Error:\n %v", err)
	}

	for _, region := range regions.Regions {
		wg.Add(1)
		go fetchInstances(sess, accountId, *region.RegionName, &wg, outputFile)
	}

	wg.Wait()
}

func fetchInstances(sess *session.Session, accountId string, region string, wg *sync.WaitGroup, outputFile os.File) {
	defer wg.Done()

	// Fetch instances:
	instancesApiCall := ec2.New(sess, aws.NewConfig().WithRegion(region))
	result, err := instancesApiCall.DescribeInstances(nil)
	if err != nil {
		log.Fatalf("Failed to call DescribeInstances. Error: \n %v", err)
	}

	for i, _ := range result.Reservations {
		for _, inst := range result.Reservations[i].Instances {
			// Fetch instance name:
			name := "None"
			for _, keys := range inst.Tags {
				if *keys.Key == "Name" {
					name = *keys.Value
				}
			}

			// Write line to output file:
			csvLine := []string{accountId, name, *inst.InstanceType, *inst.Placement.AvailabilityZone}
			writeToFile(outputFile, csvLine)
		}
	}

	log.Printf("Finished processing region %s from %s", region, accountId)
}

func writeToFile(outputFile os.File, line []string) {
	writer := csv.NewWriter(&outputFile)

	err := writer.Write(line)
	if err != nil {
		log.Fatalf("Failed to write to %s. Error:\n %v", outputFile.Name(), err)
	}

	writer.Flush()
}
