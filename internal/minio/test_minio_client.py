"""Module to test minio client class
"""

import os
import pytest
from minio_client import MinioClient
from dotenv import load_dotenv
load_dotenv()

TEST_BUCKET = "aiml-test"

@pytest.fixture(scope="module")
def minio_client() -> object:
    """Define minio client
    Returns:
        object: client instance
    """
    config = {
        'URL': os.getenv('URL'),
        'ACCESS_KEY': os.getenv('ACCESS_KEY'),
        'SECRET_KEY': os.getenv('SECRET_KEY'),
        'TLS': os.getenv('TLS') == 'True'
    }
    client = MinioClient(config)
    assert client.check_connection(), "FAILED TO CONNECT TO MINIO"
    return client

def test_create_bucket(minio_client: object):
    """Test creating minio bucket
    Args:
        minio_client (object): minio client for test
    """
    minio_client.create_bucket(TEST_BUCKET)
    assert minio_client.client.bucket_exists(TEST_BUCKET)

def test_upload_file(minio_client: object):
    """Test uploading file to bucket
    Args:
        minio_client (object): minio client for test
    """
    minio_client.upload_file(TEST_BUCKET,
                             "/home/st_yudi/personal-github-repository/aiml-infrastructure/config/config.yaml.example",
                             "config.yaml.example")
    assert minio_client.file_exists(TEST_BUCKET,
                                    "config.yaml.example")

def test_download_file(minio_client: object):
    """Test downloading file from bucket
    Args:
        minio_client (object): minio client for test
    """
    minio_client.download_file(TEST_BUCKET,
                               "config.yaml.example",
                               "config.yaml.example")
    assert os.path.exists("/home/st_yudi/personal-github-repository/aiml-infrastructure/internal/minio/config.yaml.example")
    os.remove("/home/st_yudi/personal-github-repository/aiml-infrastructure/internal/minio/config.yaml.example")

def test_list_files(minio_client: object):
    """Test listing file inside bucket
    Args:
        minio_client (object): minio client for test
    """
    minio_client.list_files(TEST_BUCKET)

def test_delete_file(minio_client: object):
    """Test delete files inside bucket
    Args:
        minio_client (object): minio client for test
    """
    minio_client.delete_file(TEST_BUCKET, "config.yaml.example")
    assert not minio_client.file_exists(TEST_BUCKET, "config.yaml.example")

def test_delete_bucket(minio_client: object):
    """Test delete minio bucket
    Args:
        mino_client (object): minio client for test
    """
    minio_client.delete_bucket(TEST_BUCKET)
    assert not minio_client.client.bucket_exists(TEST_BUCKET)