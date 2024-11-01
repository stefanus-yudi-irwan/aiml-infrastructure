from loguru import logger
from minio import Minio
from minio.error import S3Error
import urllib3

urllib3.disable_warnings()

class MinioClient():
    """Class to interact with minio
    """
    def __init__(self,
                 config: dict) -> None:
        """Initialize class.
        Args:
            config (dict): minio configuration
        """
        logger.info("INITIALIZING MINIO CLASS")
        self.config = config
        self.client = Minio(
            endpoint = self.config['URL'],
            access_key = self.config['ACCESS_KEY'],
            secret_key = self.config['SECRET_KEY'],
            secure = self.config['TLS'],
            cert_check = (not self.config['TLS']))
        
    def _error(self, function_name: str, message: str, e: Exception) -> None:
        """error wrapper function
        Args:
            function (str): function name
        """
        logger.error(f"AN ERROR OCCURED IN FUNCTION: {function_name}; MESSAGE: {message}; EXCEPTION: {e}")
        return None

    def check_connection(self) -> bool:
        """Check connection to minio server.
        Returns:
            bool: status of the connection
        """
        try:
            self.client.list_buckets()
            logger.info("Connection to minio server is succesful")
            return True
        except S3Error as e:
            self._error("check_connection","CONNECTION TO MINIO SERVER FAILED",e)
            return False
    
    def create_bucket(self, bucket_name: str) -> None:
        """Create bucket inside Minio.
        Args:
            bucket_name (str): minio bucket name
        """
        try:
            if not self.client.bucket_exists(bucket_name):
                self.client.make_bucket(bucket_name)
                logger.success(f"Bucket '{bucket_name}' created succesfully")
            else:
                logger.info(f"Bucket {bucket_name} already exists")
        except S3Error as e:
            self._error("create_bucket","CREATING BUCKET FAIL",e)
        
    def delete_bucket(self, bucket_name: str) -> None:
        """Delete bucket inside Minio.
        Args:
            bucket_name (str): minio bucket name
        """
        try:
            if self.client.bucket_exists(bucket_name):
                if self.client.remove_bucket(bucket_name):
                    logger.success(f"Bucket {bucket_name} deleted")
            else:
                logger.info(f"Bucket {bucket_name} does not exists")
        except S3Error as e:
            self._error("delete_bucket","BUCKET DELETION FAILED",e)

    def upload_file(self, bucket_name: str, file_path: str, object_name: str) -> None:
        """Upload file to a bucket.
        Args:
            bucket_name (str): minio bucket name
            file_path (str): local file path
            object_name (str): object name in minio bucket
        """
        try:
            self.client.fput_object(bucket_name, object_name, file_path)
            logger.success(f"File '{file_path}' uploaded as '{object_name}' to bucket '{bucket_name}'")
        except S3Error as e:
            self._error("upload_file", "FAILED TO UPLOAD FILE",e)
    
    def download_file(self, bucket_name: str, file_path: str, object_name: str) -> None:
        """Download a file from a bucket.
        Args:
            bucket_name (str): minio bucket name
            file_path (str): local file path
            object_name (str): object name in minio bucket
        """
        try:
            self.client.fget_object(bucket_name, object_name, file_path)
            logger.success(f"File '{object_name}' downloaded to '{file_path}'")
        except S3Error as e:
            self._error("download_file","FAILED TO DOWNLOAD FILE",e)
    
    def list_files(self, bucket_name: str):
        """List all files in a bucket.
        Args:
            bucket_name (str): minio bucket name
        """
        try:
            objects = self.client.list_objects(bucket_name)
            for obj in objects:
                logger.info(f"Found object: {obj.object_name}")
        except S3Error as e:
            self._error("list_files","FAILED TO LIST OBJECTS",e)

    def delete_file(self, bucket_name: str, object_name: str):
        """Delete a file from a bucket.
        Args:
            bucket_name (str): minio bucket name
            object_name (str): object name in minio bucket
        """
        try:
            self.client.remove_object(bucket_name, object_name)
            logger.success(f"File '{object_name}' deleted from bucket '{bucket_name}'")
        except S3Error as e:
            self._error("delete_file",f"FAILED TO DELETE FILE {object_name}",e)

    def file_exists(self, bucket_name: str, object_name: str):
        """Check if file exists in bucket
        Args:
            bucket_name (str): minio bucket name
            object_name (str): object_name in bucket
        """
        try:
            self.client.stat_object(bucket_name, object_name)
            logger.info(f"File '{object_name} exists in bucket {bucket_name}")
            return True
        except S3Error as e:
            if e.code == 'NoSuchKey':
                logger.info(f"File '{object_name} does not exists in bucket '{bucket_name}")
                return False
            else:
                self._error("check_file","ERROR CHECKING FILE EXISTENCE",e)
                return False
