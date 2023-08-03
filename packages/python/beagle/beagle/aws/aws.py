# Copyright 2023 Specter Ops, Inc.
# 
# Licensed under the Apache License, Version 2.0
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# 
# SPDX-License-Identifier: Apache-2.0

import os
import sys

from beagle.facts import CI_S3_BUCKET

try:
    import boto3
    from botocore.errorfactory import ClientError
except ImportError:
    pass

    # TODO: I got really tired of seeing this all the time. I want to make this a little better but for now I'm muting
    #       this particluar output
    #
    # print(
    #     "AWS functions disabled: boto3 and botocore python libraries are missing from PYTHONPATH",
    #     file=sys.stderr,
    # )


def s3_exists(object_name: str) -> bool:
    s3 = boto3.client("s3")

    try:
        s3.head_object(Bucket=CI_S3_BUCKET, Key=object_name)
        return True
    except ClientError as ex:
        if ex.response["Error"]["Code"] == "404":
            return False

        raise


def s3_download(object_name: str, destination_path: str) -> None:
    s3 = boto3.client("s3")
    s3.download_file(bucket=CI_S3_BUCKET, key=object_name, filename=destination_path)


def s3_upload(source_path: str, object_name: str) -> None:
    s3 = boto3.client("s3")
    s3.upload_file(bucket=CI_S3_BUCKET, key=object_name, filename=source_path)


def s3_read(object_name: str) -> bytes:
    s3 = boto3.client("s3")
    s3_object = s3.get_object(Bucket=CI_S3_BUCKET, Key=object_name)

    return s3_object["Body"].read()


def s3_write(object_name: str, output: bytes) -> None:
    s3 = boto3.client("s3")
    s3.put_object(Bucket=CI_S3_BUCKET, Key=object_name, Body=output)
