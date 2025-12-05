from netCDF4 import Dataset

nc = Dataset("argo_2019_01/nodc_D1900975_339.nc")

print("NumRecs:", nc.dimensions["N_REC"].size if "N_REC" in nc.dimensions else "No unlimited")

print("Dimensions:")
for d in nc.dimensions.values():
    print(d.name, "=", len(d))
