"""Utils module
"""
import yaml
from loguru import logger

def load_config(file_path: str) -> dict:
    """Function to load configuration file yaml
    Args:
        file_path (str): yaml file path
    Returns:
        dict: config in dict form
    """
    with open(file_path, 'r', encoding='utf-8') as config_file:
        try:
            config = yaml.safe_load(config_file)
            return config
        except yaml.YAMLError as e:
            logger.error(f"CONFIG READ ERROR: {e}")
            return None