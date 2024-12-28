import os
from base64 import b64encode

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding as asymmetric_padding
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives.padding import PKCS7


def generate_device_keys():
    """Gera o par de chaves do dispositivo"""
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


def encrypt_file(input_file_path, private_key_pem):
    """
    Encripta um arquivo usando a chave privada fornecida do dispositivo

    Args:
        input_file_path (str): Caminho do arquivo a ser encriptado
        private_key_pem (bytes): Chave privada em formato PEM

    Returns:
        dict: Dicionário contendo os dados necessários para envio à API
    """
    # Carrega a chave privada
    private_key = serialization.load_pem_private_key(private_key_pem, password=None)
    public_key = private_key.public_key()

    # Gera uma chave simétrica aleatória para AES
    symmetric_key = os.urandom(32)  # 256 bits
    iv = os.urandom(16)

    # Lê o arquivo original
    with open(input_file_path, "rb") as f:
        file_data = f.read()

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
    encrypted_symmetric_key = public_key.encrypt(
        symmetric_key,
        asymmetric_padding.OAEP(
            mgf=asymmetric_padding.MGF1(algorithm=hashes.SHA256()),
            algorithm=hashes.SHA256(),
            label=None,
        ),
    )

    # Gera a assinatura digital
    signature = private_key.sign(
        encrypted_data,
        asymmetric_padding.PSS(
            mgf=asymmetric_padding.MGF1(hashes.SHA256()),
            salt_length=asymmetric_padding.PSS.MAX_LENGTH,
        ),
        hashes.SHA256(),
    )

    try:
        public_key.verify(
            signature,
            encrypted_data,
            asymmetric_padding.PSS(
                mgf=asymmetric_padding.MGF1(hashes.SHA256()),
                salt_length=asymmetric_padding.PSS.MAX_LENGTH,
            ),
            hashes.SHA256(),
        )
        print("Assinatura válida localmente!")
    except Exception as e:
        print(f"Erro na verificação local: {e}")

    # Gera o nome do arquivo encriptado
    filename, ext = os.path.splitext(input_file_path)
    encrypted_file_path = f"{filename}_encrypted{ext}"

    # Salva o arquivo encriptado
    with open(encrypted_file_path, "wb") as f:
        f.write(encrypted_data)

    return {
        "encrypted_file_path": encrypted_file_path,
        "encrypted_data": encrypted_data,
        "encrypted_symmetric_key": b64encode(encrypted_symmetric_key).decode("utf-8"),
        "digital_signature": b64encode(signature).decode("utf-8"),
        "metadata": {"filename": os.path.basename(input_file_path)},
    }


def save_encryption_info(encrypted_symmetric_key, digital_signature, metadata):
    """
    Salva as informações de encriptação em um arquivo
    Args:
        encrypted_symmetric_key (str): Chave simétrica encriptada em base64
        digital_signature (str): Assinatura digital em base64
        metadata (dict): Metadados adicionais
    """
    with open("encryption_info.txt", "w") as f:
        f.write("INFORMAÇÕES DE ENCRIPTAÇÃO\n\n")
        f.write("encrypted_symmetric_key (base64):\n")
        f.write(encrypted_symmetric_key + "\n\n")
        f.write("digital_signature (base64):\n")
        f.write(digital_signature + "\n\n")
        if metadata:
            f.write("metadata:\n")
            for key, value in metadata.items():
                f.write(f"{key}: {value}\n")

    print("\nInformações de encriptação salvas em: encryption_info.txt")


def main():
    # Primeiro uso: gera e salva as chaves
    if not os.path.exists("device_private.pem"):
        private_key, public_key = generate_device_keys()

        # Salva as chaves
        with open("device_private.pem", "wb") as f:
            f.write(private_key)
        with open("device_public.pem", "wb") as f:
            f.write(public_key)

        print("Novas chaves geradas e salvas!")

    # Carrega a chave privada existente
    with open("device_private.pem", "rb") as f:
        private_key = f.read()

    # Arquivo a ser encriptado
    input_file = "comando.txt"

    # Cria arquivo de teste se não existir
    if not os.path.exists(input_file):
        with open(input_file, "w") as f:
            f.write("Conteúdo de teste para encriptação")
        print(f"Arquivo de teste '{input_file}' criado!")

    print(f"\nEncriptando arquivo: {input_file}")
    result = encrypt_file(input_file, private_key)

    print(f"\nArquivo encriptado salvo em: {result['encrypted_file_path']}")
    print("\nDados para API:")
    print("encrypted_symmetric_key (base64):", result["encrypted_symmetric_key"])
    print("digital_signature (base64):", result["digital_signature"])
    print("metadata:", result["metadata"])

    save_encryption_info(
        result["encrypted_symmetric_key"],
        result["digital_signature"],
        result["metadata"],
    )

    with open("device_public.pem", "r") as f:
        public_key = f.read()

        # Monta o payload
        payload = {
            "uuid": "4fa85f64-5717-4562-b3fc-2c963f66afa6",
            "ek_certificate": "string",
            "aik": "string",
            "public_key": public_key,  # O Python/requests vai fazer o escape correto automaticamente
        }
        print(payload)


if __name__ == "__main__":
    main()
