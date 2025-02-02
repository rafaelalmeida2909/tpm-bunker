import os
import random
import string


def generate_test_file(size_mb, output_path):
    """
    Generate a test file of specified size in megabytes

    Args:
        size_mb (int): Size of the file in megabytes
        output_path (str): Path where the file should be saved
    """
    # Convert MB to bytes
    size_bytes = size_mb * 1024 * 1024

    # Calculate chunk size (1MB chunks for efficient writing)
    chunk_size = 1024 * 1024

    print(f"Generating {size_mb}MB file at: {output_path}")

    try:
        with open(output_path, "wb") as f:
            remaining_bytes = size_bytes

            while remaining_bytes > 0:
                # Generate random data for this chunk
                chunk_bytes = min(chunk_size, remaining_bytes)
                chunk = "".join(
                    random.choices(string.ascii_letters + string.digits, k=chunk_bytes)
                ).encode()
                f.write(chunk)
                remaining_bytes -= chunk_bytes

                # Print progress for large files
                if size_mb >= 100:
                    progress = ((size_bytes - remaining_bytes) / size_bytes) * 100
                    print(f"Progress: {progress:.1f}%", end="\r")

        print(f"\nSuccessfully generated {size_mb}MB file")

    except Exception as e:
        print(f"Error generating file: {e}")
        if os.path.exists(output_path):
            os.remove(output_path)


def main():
    # Cria diretório de testes se não existir
    test_dir = "test_files"
    os.makedirs(test_dir, exist_ok=True)

    # Tamanhos dos arquivos em MB
    sizes = [
        1,  # 1MB
        10,  # 10MB
        50,  # 50MB
        100,  # 100MB
        250,  # 250MB
        500,  # 500MB
        1024,  # 1GB
    ]

    total_size_gb = sum(sizes) / 1024
    print(f"Atenção: Será gerado um total de {total_size_gb:.1f}GB em arquivos")
    confirm = input("Deseja continuar? (s/n): ")

    if confirm.lower() != "s":
        print("Operação cancelada")
        return

    for size in sizes:
        file_path = os.path.join(test_dir, f"teste_{size}mb.dat")
        generate_test_file(size, file_path)


if __name__ == "__main__":
    main()
