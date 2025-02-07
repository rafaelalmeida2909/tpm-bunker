import re

import matplotlib.pyplot as plt
import numpy as np
from scipy import stats


def convert_to_ms(duration_str):
    """
    Converte a duração para milissegundos baseado na unidade especificada.
    Exemplo:
    - "100.123ms" -> 100.123
    - "1.234s" -> 1234.0
    """
    if duration_str.endswith("s") and not duration_str.endswith("ms"):
        # Remove o 's' e converte para float, depois multiplica por 1000
        return float(duration_str[:-1]) * 1000
    elif duration_str.endswith("ms"):
        # Remove o 'ms' e converte para float
        return float(duration_str[:-2])
    else:
        raise ValueError(f"Formato de duração inválido: {duration_str}")


def extract_times(log_content, is_with_tpm=True):
    """
    Extrai os tempos de execução do log
    """
    encryption_times = {}
    decryption_times = {}

    # Padrões para extração com regex
    if is_with_tpm:
        pattern = r"(Encryption|Decryption) Stats - Size: (\d+\.\d+) MB, Duration: (\d+\.?\d*(?:ms|s))"
    else:
        pattern = r"(Encryption|Decryption) - Run \d+ - Size: (\d+\.\d+) MB, Duration: (\d+\.?\d*(?:ms|s))"

    for line in log_content.split("\n"):
        if not line.strip():
            continue

        match = re.search(pattern, line)
        if match:
            op_type, size, duration = match.groups()
            size = float(size)
            duration_ms = convert_to_ms(duration)  # Converte duração para ms

            if op_type == "Encryption":
                if size not in encryption_times:
                    encryption_times[size] = []
                encryption_times[size].append(duration_ms)
            else:
                if size not in decryption_times:
                    decryption_times[size] = []
                decryption_times[size].append(duration_ms)

    return encryption_times, decryption_times


def calculate_confidence_interval(data, confidence=0.95):
    """
    Calcula intervalo de confiança para um conjunto de dados
    """
    data = np.array(data)
    mean = np.mean(data)
    std_error = stats.sem(data)
    ci = stats.t.interval(confidence, len(data) - 1, loc=mean, scale=std_error)

    return {"mean": mean, "ci_lower": ci[0], "ci_upper": ci[1], "std": np.std(data)}


def analyze_performance(machine_name, with_tpm_content, without_tpm_content):
    """
    Analisa performance para uma máquina específica
    """
    # Extrai tempos
    with_tpm_enc, with_tpm_dec = extract_times(with_tpm_content, True)
    without_tpm_enc, without_tpm_dec = extract_times(without_tpm_content, False)

    results = {
        "encryption": {"with_tpm": {}, "without_tpm": {}},
        "decryption": {"with_tpm": {}, "without_tpm": {}},
    }

    # Calcula intervalos de confiança para cada tamanho de arquivo
    for size in sorted(with_tpm_enc.keys()):
        # Encryption with TPM
        results["encryption"]["with_tpm"][size] = calculate_confidence_interval(
            with_tpm_enc[size]
        )
        # Decryption with TPM
        results["decryption"]["with_tpm"][size] = calculate_confidence_interval(
            with_tpm_dec[size]
        )

        # Encryption without TPM
        results["encryption"]["without_tpm"][size] = calculate_confidence_interval(
            without_tpm_enc[size]
        )
        # Decryption without TPM
        results["decryption"]["without_tpm"][size] = calculate_confidence_interval(
            without_tpm_dec[size]
        )

    return results


def plot_performance_comparison(results, machine_name, operation_type):
    """
    Plota gráfico comparativo de tempos de execução
    """
    sizes = sorted(results[operation_type]["with_tpm"].keys())

    x = np.arange(len(sizes))
    width = 0.35

    fig, ax = plt.subplots(figsize=(12, 6))

    # Plota barras para TPM
    means_tpm = [results[operation_type]["with_tpm"][size]["mean"] for size in sizes]
    errors_tpm = [
        (
            results[operation_type]["with_tpm"][size]["ci_upper"]
            - results[operation_type]["with_tpm"][size]["ci_lower"]
        )
        / 2
        for size in sizes
    ]

    rects1 = ax.bar(
        x - width / 2,
        means_tpm,
        width,
        yerr=errors_tpm,
        label="Com TPM",
        capsize=5,
        color="#2563eb",
    )

    # Plota barras para sem TPM
    means_no_tpm = [
        results[operation_type]["without_tpm"][size]["mean"] for size in sizes
    ]
    errors_no_tpm = [
        (
            results[operation_type]["without_tpm"][size]["ci_upper"]
            - results[operation_type]["without_tpm"][size]["ci_lower"]
        )
        / 2
        for size in sizes
    ]

    rects2 = ax.bar(
        x + width / 2,
        means_no_tpm,
        width,
        yerr=errors_no_tpm,
        label="Sem TPM",
        capsize=5,
        color="#16a34a",
    )

    ax.set_ylabel("Tempo (ms)")
    ax.set_xlabel("Tamanho do Arquivo (MB)")
    ax.set_title(f"Tempo de {operation_type.capitalize()} - {machine_name}")
    ax.set_xticks(x)
    ax.set_xticklabels([f"{size} MB" for size in sizes])
    ax.legend(loc="upper left")

    plt.xticks(rotation=45)
    plt.tight_layout()
    return fig


def print_detailed_results(results, machine_name):
    """
    Imprime resultados detalhados com intervalos de confiança
    """
    print(f"\nResultados Detalhados para {machine_name}")
    print("=" * 80)

    for operation in ["encryption", "decryption"]:
        print(f"\n{operation.upper()}")
        print("-" * 80)
        print(
            f"{'Size (MB)':<10} {'Mode':<12} {'Mean (ms)':<12} {'CI Lower':<12} {'CI Upper':<12} {'Std Dev':<12}"
        )
        print("-" * 80)

        for size in sorted(results[operation]["with_tpm"].keys()):
            # Com TPM
            with_tpm = results[operation]["with_tpm"][size]
            print(
                f"{size:<10} {'Com TPM':<12} {with_tpm['mean']:<12.2f} "
                f"{with_tpm['ci_lower']:<12.2f} {with_tpm['ci_upper']:<12.2f} "
                f"{with_tpm['std']:<12.2f}"
            )

            # Sem TPM
            without_tpm = results[operation]["without_tpm"][size]
            print(
                f"{size:<10} {'Sem TPM':<12} {without_tpm['mean']:<12.2f} "
                f"{without_tpm['ci_lower']:<12.2f} {without_tpm['ci_upper']:<12.2f} "
                f"{without_tpm['std']:<12.2f}"
            )
            print("-" * 80)


def analyze_machine(machine_name, with_tpm_content, without_tpm_content):
    """
    Analisa uma máquina específica e gera resultados e gráficos
    """
    # Analisa os dados
    results = analyze_performance(machine_name, with_tpm_content, without_tpm_content)

    # Imprime resultados detalhados
    print_detailed_results(results, machine_name)

    # Gera gráficos
    enc_fig = plot_performance_comparison(results, machine_name, "encryption")
    dec_fig = plot_performance_comparison(results, machine_name, "decryption")

    return results, enc_fig, dec_fig


# Para usar o código:

# Para cada máquina
for machine_num in [1, 2, 3]:
    # Lê os arquivos
    with open(f"machine{machine_num}/result_with_tpm.txt", "r") as f:
        with_tpm_content = f.read()
    with open(f"machine{machine_num}/result_without_tpm.txt", "r") as f:
        without_tpm_content = f.read()

    # Analisa a máquina
    results, enc_fig, dec_fig = analyze_machine(
        f"Machine {machine_num}", with_tpm_content, without_tpm_content
    )

    # Salva os gráficos
    enc_fig.savefig(f"machine{machine_num}_encryption_times.png")
    dec_fig.savefig(f"machine{machine_num}_decryption_times.png")
    plt.close("all")
