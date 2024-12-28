import hashlib
from base64 import b64encode


def verify_file_integrity(file_path):
    with open(file_path, "rb") as f:
        data = f.read()

    # Imprime tamanho e alguns detalhes dos dados
    print(f"Tamanho do arquivo: {len(data)} bytes")
    print(f"Primeiros 16 bytes (IV): {data[:16].hex()}")
    print(f"SHA256: {hashlib.sha256(data).hexdigest()}")

    # Verifica se o tamanho é múltiplo de 16 (bloco AES)
    print(f"É múltiplo de 16: {len(data) % 16 == 0}")
    return data


print("=== Arquivo Original ===")
original_data = verify_file_integrity("comando_encriptado.txt")

print("\n=== Arquivo Baixado da API ===")
downloaded_data = verify_file_integrity("encrypted_data.bin")

print("\n=== Comparação ===")
print(f"Arquivos são idênticos: {original_data == downloaded_data}")
if original_data != downloaded_data:
    # Encontra onde os dados começam a diferir
    for i, (a, b) in enumerate(zip(original_data, downloaded_data)):
        if a != b:
            print(f"Primeira diferença no byte {i}")
            print(f"Original: {a:02x}, Baixado: {b:02x}")
            break
