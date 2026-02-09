package utils

var DataDir = GetEnv("DataDir", "data/")

var RegistryFile = DataDir + "/models.json"
var SQLiteFile = DataDir + "/xiacutai.db"
var StorageDir = DataDir + "/storage/"
