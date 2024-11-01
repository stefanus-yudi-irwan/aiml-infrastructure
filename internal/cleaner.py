"""DataCleaner class
"""
import os
import pickle
import pandas as pd
from loguru import logger

DataCleanerConfig = {
    "RAW_DATA_PATH ":"",
    "CLEAN_DATA_DIR":"",
    "STOPWORDS_DIR": "",
    "FREQUENTWORDS_DIR": "",
    "LABEL_ENCODER_DIR": "",
    "ONE_HOT_ENCODER_DIR":"",
    "LOGGER_DIR": "",
    "TEST_DATA_RATIO": 0.2,
}

class DataCleaner():
    """class to provide clean data for modelling
    """
    def __init__(self,
                 config: dict) -> None:
        """data cleaner class initialization
        Args:
            config (dict): data cleaner configuration
        Returns:
            None
        """
        # save configuration to class
        self.dc_config = config
        
        # initialize logger
        self.dc_logger = logger.bind(task="DataCleaner")
        logger.add(self.dc_config["LOGGER_DIR"] + "DataCleaner.log",
                   rotation = '50 MB',
                   filter=lambda record: record['extra']['task'] == "DataCleaner")
        
        
    def get_clean_data(self, cleaning: bool=False) -> pd.DataFrame:
        """get clean data from environment
        Args:
            cleaning (bool):
                False for not starting the clean process
                True for starting the clean process
        Returns:
            train_data (pd.DataFrame): cleaned train data
            val_data (pd.DataFrame): cleaned val data
            test_data (pd.DataFrame): uncleaned test data
        """
        # load clean data if exists in clean directory
        if (os.path.exists(self.dc_config['CLEAN_DATA_DIR'])
            & (len(os.listdir(self.dc_config['CLEAN_DATA_DIR']))==3)
            & (not cleaning)):
            train_data, val_data, test_data = self.load_clean_data()
            
        # process raw data if not exists
        else:
            train_data, val_data, test_data = self.clean_data()
            
        return train_data, val_data, test_data
    
    def load_clean_data(self):
        """load clean data from clean dir in dataframe
        Args:
            None
        Returns:
            train_data (pd.DataFrame): cleaned train data
            val_data (pd.DataFrame): cleaned val data
            test_data (pd.DataFrame): uncleaned test data
        """
        self.dc_logger.info("BEGIN LOADING TRAIN DATA")
        
        self.dc_logger.info("LOAD CLEAN TRAIN DATA")
        # load train_data
        with open(f"{self.dc_config['CLEAN_DATA_DIR']}/train.pkl", "rb") as train_file:
            train_data = pickle.load(train_file)
        self.dc_logger.info("CLEAN TRAIN DATA LOADED")
        
        self.dc_logger.info("LOAD CLEAN VAL DATA")
        # load val_data
        with open(f"{self.dc_config['CLEAN_DATA_DIR']}/val.pkl", "rb") as val_file:
            val_data = pickle.load(val_file)
        self.dc_logger.info("CLEAN VAL DATA LOADED")
            
        # load test_data
        self.dc_logger.info("LOAD TEST DATA")
        with open(f"{self.dc_config['CLEAN_DATA_DIR']}/test.pkl", "rb") as test_file:
            test_data = pickle.load(test_file)
        self.dc_logger.info("TEST DATA LOADED")
        
        self.dc_logger.success("LOADING DATA COMPLETE")
            
        return train_data, val_data, test_data
        
    def clean_data(self):
        """process raw data into clean data and save in
           clean data directory
        Args: 
            None
        Returns:
            train_data (pd.DataFrame): cleaned train data
            val_data (pd.DataFrame): cleaned val data
            test_data (pd.DataFrame): uncleaned test data
        """
        self.dc_logger.info("BEGIN DATA CLEANING PROCESS")
        
        # load raw data
        self.dc_logger.info(f"LOAD RAW DATA FROM: {self.dc_config['RAW_DATA_DIR']}")
        self.load_raw_data()
        self.dc_logger.info("RAW DATA LOADED")
        
        # split train, val, test data
        self.dc_logger.info("SPLIT TRAIN, VAL, AND TEST DATA")
        train_data, val_data, test_data = self.train_test_split()
        self.dc_logger.info("DATA SPLITTED")
        
        # clean train data
        self.dc_logger.info("CLEAN TRAIN DATA")
        train_data = self.clean_raw_data(train_data, mode='train')
        self.dc_logger.info("TRAIN DATA CLEANED")
        
        # clean val data
        self.dc_logger.info("CLEAN VAL DATA")
        val_data = self.clean_raw_data(val_data, mode='val')
        self.dc_logger.info("VAL DATA CLEANED")
        
        # save train, val, and test data
        self.dc_logger.info(f"SAVE TRAIN, VAL, AND TEST DATA IN: {self.dc_config['CLEAN_DATA_DIR']}")
        self.save_clean_data(train_data, val_data, test_data)
        self.dc_logger.info(f"TRAIN, VAL, AND TEST DATA SAVED")
        
        self.dc_logger.success("CLEANING DATA COMPLETE")

        return train_data, val_data, test_data
        
    def load_raw_data(self):
        """load raw train, val, and test data
        Args: <flexible>
        Returns: <flexible>
        """
        pass
    
    def train_test_split(self, data: pd.DataFrame) -> pd.DataFrame:
        
        pass
        
    def clean_raw_data(self, data):
        """clean raw textual data
        Args: <flexible>
        Returns: <flexible>
        """
        
        def handle_html(text_input: str) -> str:
            pass
        
        def handle_stopwords(text_input: str) -> str:
            pass
        
        def handle_punctuation(text_input: str) -> str:
            pass
        
        def handle_number(text_input: str) -> str:
            pass
        
        def handle_pos_tag(text_input: str) -> str:
            pass
        
        def handle_textual_normalization(text_input: str) -> str:
            pass
        
        def handle_frequent_words(text_input: str) -> str:
            pass
        
        def handle_white_space(text_input: str) -> str:
            pass
        
        def handle_label_encoder(text_input: str) -> str:
            pass
        
        def handle_one_hot_encoder(text_input: str) -> str:
            pass
        
        def handle_imbalance(text_input: str) -> str:
            pass

        def handle_special_token(text_input: str) -> str:
            pass 

        pass
    
    def save_clean_data(self, train_data: pd.DataFrame, val_data: pd.DataFrame, test_data: pd.DataFrame) -> None:
        """save clean data into clean dir in pkl file
        Args:
            train_data (pd.DataFrame): cleaned train data
            val_data (pd.DataFrame): cleaned val data
            test_data (pd.DataFrame): uncleaned test data
        Returns:
            None
        """
        # make directory to save data if not exists
        if not os.path.exists(self.dc_config['CLEAN_DATA_DIR']):
            os.makedirs(self.dc_config['CLEAN_DATA_DIR'])

        # save train_data
        with open(f"{self.dc_config['CLEAN_DATA_DIR']}/train.pkl", "wb") as train_file:
            pickle.dump(train_data, train_file)
            
        # save val_data
        with open(f"{self.dc_config['CLEAN_DATA_DIR']}/val.pkl", "wb") as val_file:
            pickle.dump(val_data, val_file)
            
        # save test_data
        with open(f"{self.dc_config['CLEAN_DATA_DIR']}/test.pkl", "wb") as test_file:
            pickle.dump(test_data, test_file)