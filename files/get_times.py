with open("stats_without_tpm.txt", "r") as f:
    linhas = f.readlines()
    stats_encrypt = []
    stats_decrypt = []
    for l in linhas:
        if "Encryption" in l:
            stats_encrypt.append(l)
        if "Decryption" in l:
            stats_decrypt.append(l)

for s in stats_encrypt:
    print(s)

for s in stats_decrypt:
    print(s)
