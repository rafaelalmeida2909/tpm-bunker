'use client';

import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export default function TPMGraphs() {
  const [data, setData] = useState({
    machine1: { withTPM: null, withoutTPM: null },
    machine2: { withTPM: null, withoutTPM: null },
    machine3: { withTPM: null, withoutTPM: null }
  });

  const processFileContent = (content) => {
    const lines = content.split('\n');
    const encryptionData = {};
    const decryptionData = {};
    let currentSection = null;
    
    lines.forEach((line) => {
      if (line.includes('Encriptação:') || line.includes('Encryption')) {
        currentSection = 'encryption';
      } else if (line.includes('Decriptação:') || line.includes('Decryption')) {
        currentSection = 'decryption';
      }
      
      // More flexible regex to match different log formats
      const statsRegex = /Size:\s*([\d.]+)\s*MB.*?Duration:\s*([\d.]+)\s*(ms|s)/i;
      const matches = line.match(statsRegex);
      
      if (matches) {
        const size = parseFloat(matches[1]);
        let duration = parseFloat(matches[2]);
        if (matches[3].toLowerCase() === 's') {
          duration *= 1000; // Convert seconds to milliseconds
        }
        
        const section = currentSection === 'encryption' ? encryptionData : decryptionData;
        if (!section[size]) {
          section[size] = [];
        }
        section[size].push(duration);
      }
    });

    // Calculate averages
    const calculateAverages = (data) => {
      return Object.entries(data).map(([size, durations]) => ({
        size: parseFloat(size),
        duration: durations.reduce((a, b) => a + b, 0) / durations.length
      })).sort((a, b) => a.size - b.size);
    };

    const result = {
      encryption: calculateAverages(encryptionData),
      decryption: calculateAverages(decryptionData)
    };

    return result;
  };

  useEffect(() => {
    const loadData = async () => {
      try {
        const machines = ['machine1', 'machine2', 'machine3'];
        const results = {};

        for (const machine of machines) {
          results[machine] = {};
          
          try {
            // Load with TPM results
            console.log(`Fetching ${machine}/result_with_tpm.txt`);
            const withTPMResponse = await fetch(`/${machine}/result_with_tpm.txt`);
            if (!withTPMResponse.ok) {
              throw new Error(`HTTP error! status: ${withTPMResponse.status} for with_tpm`);
            }
            const withTPMContent = await withTPMResponse.text();
            console.log(`${machine} with TPM content loaded, length:`, withTPMContent.length);
            
            // Load without TPM results
            console.log(`Fetching ${machine}/result_without_tpm.txt`);
            const withoutTPMResponse = await fetch(`/${machine}/result_without_tpm.txt`);
            if (!withoutTPMResponse.ok) {
              throw new Error(`HTTP error! status: ${withoutTPMResponse.status} for without_tpm`);
            }
            const withoutTPMContent = await withoutTPMResponse.text();
            console.log(`${machine} without TPM content loaded, length:`, withoutTPMContent.length);

            // Process both contents
            const withTPMData = processFileContent(withTPMContent);
            const withoutTPMData = processFileContent(withoutTPMContent);

            console.log(`${machine} processed data:`, {
              withTPM: withTPMData,
              withoutTPM: withoutTPMData
            });

            results[machine] = {
              withTPM: withTPMData,
              withoutTPM: withoutTPMData
            };

          } catch (error) {
            console.error(`Error loading data for ${machine}:`, error);
          }
        }

        setData(results);
      } catch (error) {
        console.error('Error loading data:', error);
      }
    };

    loadData();
  }, []);

  const renderChart = (machineData, operation, machineName) => {
    if (!machineData?.withTPM?.[operation]) {
      console.log(`Missing withTPM data for ${machineName} - ${operation}`);
      return null;
    }
    
    if (!machineData?.withoutTPM?.[operation]) {
      console.log(`Missing withoutTPM data for ${machineName} - ${operation}`);
      return null;
    }

    // Get all unique sizes
    const allSizes = [...new Set([
      ...machineData.withTPM[operation].map(d => d.size),
      ...machineData.withoutTPM[operation].map(d => d.size)
    ])].sort((a, b) => a - b);

    const chartData = allSizes.map(size => {
      const withTPMData = machineData.withTPM[operation].find(d => d.size === size);
      const withoutTPMData = machineData.withoutTPM[operation].find(d => d.size === size);
      
      return {
        size,
        'Com TPM': withTPMData?.duration,
        'Sem TPM': withoutTPMData?.duration
      };
    });

    const formatYAxis = (value) => {
      if (value >= 1000) {
        return `${(value / 1000).toFixed(1)}s`;
      }
      return `${value.toFixed(0)}ms`;
    };

    return (
      <div className="w-full mb-8 p-4 bg-white rounded-lg shadow">
        <h2 className="text-xl font-bold mb-4">
          {`${machineName} - ${operation === 'encryption' ? 'Criptografia' : 'Descriptografia'}`}
        </h2>
        <div className="h-[400px]">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="size" 
                label={{ value: 'Tamanho (MB)', position: 'insideBottom', offset: -5 }} 
              />
              <YAxis 
                label={{ value: 'Duração', angle: -90, position: 'insideLeft', offset: 10 }}
                tickFormatter={formatYAxis}
              />
              <Tooltip 
                formatter={(value) => [`${value?.toFixed(2)}ms`, null]}
              />
              <Legend 
                align="left"
                verticalAlign="top"
                height={36}
              />
              <Line 
                type="monotone" 
                dataKey="Com TPM" 
                stroke="#2563eb" 
                strokeWidth={2} 
                dot={{ r: 4 }} 
                connectNulls 
              />
              <Line 
                type="monotone" 
                dataKey="Sem TPM" 
                stroke="#16a34a" 
                strokeWidth={2} 
                dot={{ r: 4 }} 
                connectNulls 
              />
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
        <div key={machine} className="space-y-6">
          {renderChart(machineData, 'encryption', machine)}
          {renderChart(machineData, 'decryption', machine)}
        </div>
      ))}
      <div className="mt-4 p-4 bg-gray-100 rounded">
        <h3 className="font-bold">Debug Info:</h3>
        <pre className="mt-2 text-sm overflow-auto">
          {JSON.stringify(data, null, 2)}
        </pre>
      </div>
    </div>
  );
}