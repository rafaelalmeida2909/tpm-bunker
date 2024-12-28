import os
from base64 import b64decode

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives.padding import PKCS7


def decrypt_file(encrypted_file_path, private_key_path, encrypted_symmetric_key):
    # Carrega a chave privada
    with open(private_key_path, "rb") as key_file:
        private_key = serialization.load_pem_private_key(key_file.read(), password=None)

    # Lê o arquivo encriptado
    with open(encrypted_file_path, "rb") as f:
        encrypted_data = f.read()

    # Separa o IV e o conteúdo encriptado
    iv = encrypted_data[:16]
    encrypted_content = encrypted_data[16:]

    try:
        # Decripta a chave simétrica
        symmetric_key = private_key.decrypt(
            encrypted_symmetric_key,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None,
            ),
        )

        # Decripta o conteúdo
        cipher = Cipher(algorithms.AES(symmetric_key), modes.CBC(iv))
        decryptor = cipher.decryptor()
        padded_data = decryptor.update(encrypted_content) + decryptor.finalize()

        # Remove o padding
        unpadder = PKCS7(128).unpadder()
        decrypted_data = unpadder.update(padded_data) + unpadder.finalize()

        return decrypted_data

    except Exception as e:
        raise Exception(f"Erro na decriptação: {str(e)}")


def main():
    encrypted_file = "encrypted_data.bin"
    output_file = "comando_decriptado.txt"
    private_key_file = "device_private.pem"

    if not os.path.exists(encrypted_file):
        print(f"Arquivo encriptado '{encrypted_file}' não encontrado!")
        return

    if not os.path.exists(private_key_file):
        print(f"Arquivo de chave privada '{private_key_file}' não encontrado!")
        return

    try:
        # Lê a chave simétrica encriptada
        with open("encryption_info.txt", "r") as f:
            content = f.read()
            for line in content.split("\n"):
                if line.startswith("encrypted_symmetric_key (base64):"):
                    encrypted_symmetric_key = b64decode(
                        content.split("\n")[content.split("\n").index(line) + 1].strip()
                    )
                    break
            else:
                raise Exception(
                    "Chave simétrica não encontrada no arquivo de informações"
                )

        # Decripta o arquivo
        decrypted_data = decrypt_file(
            encrypted_file, private_key_file, encrypted_symmetric_key
        )

        # Salva o arquivo decriptado
        with open(output_file, "wb") as f:
            f.write(decrypted_data)

        print(f"\nArquivo decriptado com sucesso!")
        print(f"Arquivo original decriptado salvo como: {output_file}")

    except Exception as e:
        print(f"Erro durante a decriptação: {str(e)}")


if __name__ == "__main__":

    main()
