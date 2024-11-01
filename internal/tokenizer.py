"""tokenizer class
"""

import pandas as pd
from loguru import logger

DataTokenizerConfig = {
    "BATCH_SIZE":32,
    "BUFFER_SIZE":64,
    "INPUT_SEQ_LENGTH":150,
    "TARGET_SEQ_LENGTH":50,
    "LOGGER_DIR":"",
    "VOCAB_DIR":"",
    "VECTORIZER_DIR":""
}

class DataTokenizer():
    """class to tokenize textual data
    """
    def __init__(self,
                 config: dict) -> None:
        """
        """
        # save configuration to class
        self.dt_config = config
        
        # initialize logger
        self.dt_logger = logger.bind(task="DataTokenizer")
        logger.add(self.dt_config['LOGGER_DIR'] + "DataTokenizer.log",
                   rotation = '50 MB',
                   filter=lambda record: record['extra']['task'] == "DataTokenizer")
        
        self.dt_logger.info("INITIALIZE DATA TOKENIZER CLASS")

    def get_data_tokenized(train_data: pd.DataFrame, val_data: pd.DataFrame):

        # create internal class tokenizer
        self.create_tokenizer()

        # create dataset
        train_dataset, val_dataset = self.create_dataset(train_data, val_data)
        pass
        
    def create_tokenizer(self):
        self.dt_logger.info("BEGIN CREATE TOKENIZER")

        self.dt_logger.info("FINISH CREATE TOKENIZER")
        pass

    def create_dataset(self, train_data: pd.DataFrame, val_data: pd.DataFrame):
        self.dt_logger.info("BEGIN CREATE TRAIN AND VAL DATASET")

        self.dt_logger.info("FINISH CREATE TRAIN AND VAL DATASET")
        pass

    def save_vocabulary(self):
        self.dt_logger.info(f"SAVING VOCABULARY TO: {self.dt_config['VOCAB_DIR']}/vocab.txt")

        self.dt_logger.info("VOCABULARY SAVED")
        pass

    def load_vocabulary(self):
        self.dt_logger.info(f"LOADING VOCABULARY FROM: {self.dt_config['VOCAB_DIR']}/vocab.txt")

        self.dt_logger.info("VOCABULARY LOADED")
        pass

    def save_tokenizer(self):
        pass

    def load_tokenizer(self):
        pass
        
    def tokenize_data(self):
        pass
    
    def detokenize_data(self):
        pass

    def batch_data(self):
        pass
    
    def pad_data(self):
        pass

