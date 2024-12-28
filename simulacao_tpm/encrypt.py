import os
from base64 import b64encode

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding as asymmetric_padding
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives.padding import PKCS7


def generate_key_pair():
    private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    public_key = private_key.public_key()

    private_pem = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    )
    public_pem = public_key.public_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PublicFormat.SubjectPublicKeyInfo,
    )

    return private_pem, public_pem


def encrypt_file(input_file_path):
    # Gera o par de chaves
    private_key, public_key = generate_key_pair()

    # Gera uma chave simétrica aleatória para AES
    symmetric_key = os.urandom(32)  # 256 bits

    # Gera um IV aleatório para AES
    iv = os.urandom(16)

    # Lê o arquivo original
    with open(input_file_path, "rb") as f:
        file_data = f.read()

    # Calcula o hash do arquivo original
    hasher = hashes.Hash(hashes.SHA256())
    hasher.update(file_data)
    hash_original = hasher.finalize()

    # Adiciona padding PKCS7
    padder = PKCS7(128).padder()
    padded_data = padder.update(file_data) + padder.finalize()

    # Encripta os dados com AES
    cipher = Cipher(algorithms.AES(symmetric_key), modes.CBC(iv))
    encryptor = cipher.encryptor()
    encrypted_content = encryptor.update(padded_data) + encryptor.finalize()

    # Combina IV com o conteúdo encriptado
    encrypted_data = iv + encrypted_content

    # Encripta a chave simétrica com RSA
    public_key_loaded = serialization.load_pem_public_key(public_key)
    encrypted_symmetric_key = public_key_loaded.encrypt(
        symmetric_key,
        asymmetric_padding.OAEP(
            mgf=asymmetric_padding.MGF1(algorithm=hashes.SHA256()),
            algorithm=hashes.SHA256(),
            label=None,
        ),
    )

    # Gera a assinatura digital
    private_key_loaded = serialization.load_pem_private_key(private_key, password=None)
    signature = private_key_loaded.sign(
        encrypted_data,
        asymmetric_padding.PSS(
            mgf=asymmetric_padding.MGF1(hashes.SHA256()),
            salt_length=asymmetric_padding.PSS.MAX_LENGTH,
        ),
        hashes.SHA256(),
    )

    # Gera o nome do arquivo encriptado
    filename, ext = os.path.splitext(input_file_path)
    encrypted_file_path = f"{filename}_encriptado{ext}"

    # Salva o arquivo encriptado
    with open(encrypted_file_path, "wb") as f:
        f.write(encrypted_data)

    return {
        "encrypted_file_path": encrypted_file_path,
        "encrypted_data": encrypted_data,
        "encrypted_symmetric_key": encrypted_symmetric_key,
        "digital_signature": signature,
        "hash_original": hash_original,
        "public_key": public_key,
        "private_key": private_key,
    }




def main():
    # Arquivo a ser encriptado
    input_file = "comando.txt"

    if not os.path.exists(input_file):
        with open(input_file, "w") as f:
            f.write("Conteúdo de teste para encriptação")
        print(f"Arquivo de teste '{input_file}' criado!")

    print(f"\nEncriptando arquivo: {input_file}")
    result = encrypt_file(input_file)

    print(f"\nArquivo encriptado salvo em: {result['encrypted_file_path']}")


    # Salva as informações de encriptação
    with open("encryption_info.txt", "w") as f:
        f.write("DADOS PARA API:\n\n")
        f.write("encrypted_symmetric_key (base64):\n")
        f.write(b64encode(result["encrypted_symmetric_key"]).decode("utf-8") + "\n\n")
        f.write("digital_signature (base64):\n")
        f.write(b64encode(result["digital_signature"]).decode("utf-8") + "\n\n")
        f.write("hash_original (base64):\n")
        f.write(b64encode(result["hash_original"]).decode("utf-8"))

    # Salva as chaves
    with open("public_key.pem", "wb") as f:
        f.write(result["public_key"])

    with open("private_key.pem", "wb") as f:
        f.write(result["private_key"])

    print("\nArquivos gerados:")
    print(f"- {os.path.basename(result['encrypted_file_path'])} (arquivo encriptado)")
    print("- encryption_info.txt (dados para API)")
    print("- public_key.pem (chave pública)")
    print("- private_key.pem (chave privada)")
    print("\nATENÇÃO: Mantenha o arquivo private_key.pem seguro!")


if __name__ == "__main__":
    main()
