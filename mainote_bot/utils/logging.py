import logging
import sys

def setup_logging():
    """Configure logging to stdout for fly.io"""
    logging.basicConfig(
        stream=sys.stdout,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        level=logging.INFO
    )
    return logging.getLogger(__name__)

# Create a logger for this module
logger = setup_logging()