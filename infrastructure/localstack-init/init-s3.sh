#!/bin/bash
echo "🪣 Initializing Local S3 Buckets via LocalStack..."
awslocal s3 mb s3://ems-tickets
echo "✅ Local S3 Bucket 'ems-tickets' successfully created."