'use client';

import { useEffect, useState } from 'react';
import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

export default function TPMGraphs() {
  const [data, setData] = useState({});

  const processFileContent = (content) => {
    const lines = content.split('\n');
    const result = { encryption: {}, decryption: {} };
    let currentSection = null;

    lines.forEach((line) => {
      if (line.includes('Encriptação:') || line.includes('Encryption')) {
        currentSection = 'encryption';
      } else if (line.includes('Decriptação:') || line.includes('Decryption')) {
        currentSection = 'decryption';
      }

      const statsRegex = /Size:\s*([\d.]+)\s*MB.*?Duration:\s*([\d.]+)\s*(ms|s)/i;
      const matches = line.match(statsRegex);

      if (matches) {
        const size = parseFloat(matches[1]);
        let duration = parseFloat(matches[2]);
        if (matches[3].toLowerCase() === 's') duration *= 1000;

        if (!result[currentSection][size]) result[currentSection][size] = [];
        result[currentSection][size].push(duration);
      }
    });

    return Object.keys(result).reduce((acc, key) => {
      acc[key] = Object.entries(result[key]).map(([size, durations]) => ({
        size: parseFloat(size),
        duration: durations.reduce((a, b) => a + b, 0) / durations.length
      })).sort((a, b) => a.size - b.size);
      return acc;
    }, {});
  };

  useEffect(() => {
    const loadData = async () => {
      const machines = ['machine1', 'machine2', 'machine3'];
      const results = {};

      for (const machine of machines) {
        results[machine] = {};
        try {
          const withTPMResponse = await fetch(`/${machine}/result_with_tpm.txt`);
          const withoutTPMResponse = await fetch(`/${machine}/result_without_tpm.txt`);
          if (!withTPMResponse.ok || !withoutTPMResponse.ok) throw new Error('Erro ao buscar arquivos');

          const withTPMData = processFileContent(await withTPMResponse.text());
          const withoutTPMData = processFileContent(await withoutTPMResponse.text());

          results[machine] = { withTPM: withTPMData, withoutTPM: withoutTPMData };
        } catch (error) {
          console.error(`Erro ao carregar dados para ${machine}:`, error);
        }
      }
      setData(results);
    };

    loadData();
  }, []);

  const renderChart = (machineData, machineName) => {
    if (!machineData.withTPM || !machineData.withoutTPM) return null;
    
    const allSizes = [...new Set([
      ...machineData.withTPM.encryption.map(d => d.size),
      ...machineData.withTPM.decryption.map(d => d.size),
      ...machineData.withoutTPM.encryption.map(d => d.size),
      ...machineData.withoutTPM.decryption.map(d => d.size)
    ])].sort((a, b) => a - b);

    const chartData = allSizes.map(size => {
      const findData = (operation, source) => machineData[source]?.[operation]?.find(d => d.size === size)?.duration;
      return {
        size,
        'Criptografia - Com TPM': findData('encryption', 'withTPM'),
        'Criptografia - Sem TPM': findData('encryption', 'withoutTPM'),
        'Descriptografia - Com TPM': findData('decryption', 'withTPM'),
        'Descriptografia - Sem TPM': findData('decryption', 'withoutTPM')
      };
    });

    return (
      <div className="w-full mb-8 p-4 bg-white rounded-lg shadow">
        <h2 className="text-xl font-bold mb-4">{machineName} - Comparação TPM</h2>
        <div className="h-[400px]">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="size" label={{ value: 'Tamanho (MB)', position: 'insideBottom', offset: -5, fill: '#000' }} tick={{ fill: '#000' } } />
              <YAxis label={{ value: 'Duração (ms)', angle: -90, position: 'insideLeft', offset: 5, fill: '#000' }} tick={{ fill: '#000' } } />
              <Tooltip formatter={(value) => [`${value?.toFixed(2)}ms`, null]} />
              <Legend align="left" verticalAlign="top" height={36} />
              <Line type="monotone" dataKey="Criptografia - Com TPM" stroke="blue" strokeWidth={3} dot={{ r: 4 }} />
              <Line type="monotone" dataKey="Criptografia - Sem TPM" stroke="#60a5fa" strokeWidth={3} dot={{ r: 4 }} />
              <Line type="monotone" dataKey="Descriptografia - Com TPM" stroke="red" strokeWidth={3} dot={{ r: 4 }} />
              <Line type="monotone" dataKey="Descriptografia - Sem TPM" stroke="#f87171" strokeWidth={3} dot={{ r: 4 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    );
  };

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h1 className="text-2xl font-bold mb-6">Comparação de Performance TPM</h1>
      {Object.entries(data).map(([machine, machineData]) => (
        <div key={machine}>{renderChart(machineData, machine)}</div>
      ))}
      <div className="mt-4 p-4 bg-gray-100 rounded">
        <h3 className="font-bold">Debug Info:</h3>
        <pre className="mt-2 text-sm overflow-auto">{JSON.stringify(data, null, 2)}</pre>
      </div>
    </div>
  );
}
