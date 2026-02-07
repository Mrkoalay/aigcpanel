package utils

var DATA_DIR = GetEnv("DATA_DIR", "data/")

var RegistryFile = DATA_DIR + "/models.json"
var SQLiteFile = DATA_DIR + "/xiacutai.db"
