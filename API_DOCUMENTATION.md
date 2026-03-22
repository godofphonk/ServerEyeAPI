# ServerEye API Documentation

**Version:** 1.0.0  
**Base URL:** `https://api.servereye.com` (production) или `http://localhost:8080` (development)

---

## 📋 Содержание

1. [Введение](#введение)
2. [Аутентификация](#аутентификация)
3. [Эндпоинты метрик](#эндпоинты-метрик)
4. [Эндпоинты статуса](#эндпоинты-статуса)
5. [Эндпоинты статической информации](#эндпоинты-статической-информации)
6. [Эндпоинты истории метрик](#эндпоинты-истории-метрик)
7. [Эндпоинты алертов](#эндпоинты-алертов)
8. [Эндпоинты управления серверами](#эндпоинты-управления-серверами)
9. [Коды ошибок](#коды-ошибок)
10. [Примеры использования](#примеры-использования)

---

## Введение

ServerEye API предоставляет REST интерфейс для мониторинга серверов в реальном времени. API позволяет получать метрики производительности, статус серверов, историю данных и управлять алертами.

### Основные возможности:

- ✅ Получение метрик производительности в реальном времени
- ✅ Мониторинг статуса серверов (online/offline)
- ✅ Доступ к статической информации о серверах
- ✅ История метрик с автоматической гранулярностью
- ✅ Система алертов и уведомлений
- ✅ Управление источниками данных (Telegram, Discord и т.д.)

---

## Аутентификация

API использует два типа аутентификации:

### 1. Server Key (для агентов)

Используется агентами для отправки метрик.

```
Header: X-Server-Key: key_xxxxx
```

### 2. API Key (для веб-приложений)

Используется веб-приложениями для чтения данных.

```
Header: Authorization: Bearer your_api_key
```

### Получение ключей:

```http
POST /api/admin/keys
Content-Type: application/json

{
  "description": "My Web Application",
  "permissions": ["read:metrics", "read:servers"]
}
```

---

## Эндпоинты метрик

### 1. Получить текущие метрики сервера

Возвращает последние динамические метрики сервера.

**Эндпоинт:**
```
GET /api/servers/by-key/{server_key}/metrics
GET /api/servers/{server_id}/metrics
```

**Параметры:**
- `server_key` (string) - Ключ сервера
- `server_id` (string) - ID сервера

**Ответ:**
```json
{
  "server_id": "srv_38822172",
  "server_key": "key_f1a5af38",
  "metrics": {
    "cpu_percent": 25.5,
    "memory_percent": 65.2,
    "disk_percent": 45.1,
    "network_mbps": 3.5,
    "load_average": {
      "1m": 0.82,
      "5m": 1.59,
      "15m": 1.75
    },
    "temperature_celsius": 48.5,
    "temperatures": {
      "cpu": 31.0,
      "gpu": 48.5,
      "storage": [
        {
          "device": "/dev/nvme0n1",
          "temperature": 29.85
        }
      ],
      "highest": 48.5
    },
    "processes_total": 572,
    "processes_running": 1,
    "processes_sleeping": 571,
    "uptime_seconds": 9563,
    "memory_details": {
      "used_gb": 13.39,
      "available_gb": 16.58,
      "free_gb": 4.33,
      "buffers_gb": 0.64,
      "cached_gb": 11.94
    },
    "disk_details": [
      {
        "path": "/",
        "used_gb": 428,
        "free_gb": 97,
        "used_percent": 82
      }
    ],
    "network_details": {
      "total_rx_mbps": 0.36,
      "total_tx_mbps": 0.49
    },
    "timestamp": "2026-03-22T12:21:19Z"
  }
}
```

**Пример запроса (JavaScript):**
```javascript
const response = await fetch('https://api.servereye.com/api/servers/by-key/key_f1a5af38/metrics', {
  headers: {
    'Authorization': 'Bearer your_api_key'
  }
});
const data = await response.json();
console.log(`CPU: ${data.metrics.cpu_percent}%`);
```

**Пример запроса (cURL):**
```bash
curl -H "Authorization: Bearer your_api_key" \
  https://api.servereye.com/api/servers/by-key/key_f1a5af38/metrics
```

---

### 2. Получить метрики с температурами

Возвращает метрики с детальной информацией о температурах.

**Эндпоинт:**
```
GET /api/servers/by-key/{server_key}/metrics/temperatures
GET /api/servers/{server_id}/metrics/temperatures
```

**Ответ:**
```json
{
  "server": {
    "id": "srv_38822172",
    "name": "web-server-01",
    "status": "online"
  },
  "metrics": {
    "basic": {
      "cpu": 25.5,
      "memory": 65.2,
      "disk": 45.1,
      "network": 3.5
    },
    "temperature": {
      "cpu": 31,
      "gpu": 48,
      "system": 0,
      "highest": 48,
      "unit": "celsius",
      "storage": [
        {
          "device": "/dev/nvme0n1",
          "temperature": 29.85,
          "status": "normal",
          "threshold": 75,
          "severity": "info",
          "message": "Storage temperature normal"
        }
      ]
    },
    "timestamp": 1774182079
  }
}
```

---

## Эндпоинты статуса

### 1. Получить статус сервера

Возвращает только статус сервера без метрик.

**Эндпоинт:**
```
GET /api/servers/by-key/{server_key}/status
GET /api/servers/{server_id}/status
```

**Ответ:**
```json
{
  "server_id": "srv_38822172",
  "server_key": "key_f1a5af38",
  "online": true,
  "last_seen": "2026-03-22T12:21:03Z",
  "agent_version": "1.2.1"
}
```

**Описание полей:**
- `online` (boolean) - Сервер онлайн (last_seen < 5 минут)
- `last_seen` (timestamp) - Время последнего получения метрик
- `agent_version` (string) - Версия агента на сервере

**Пример использования:**
```javascript
async function checkServerStatus(serverKey) {
  const response = await fetch(`/api/servers/by-key/${serverKey}/status`);
  const status = await response.json();
  
  if (status.online) {
    console.log('✅ Server is online');
  } else {
    console.log('❌ Server is offline');
    console.log(`Last seen: ${status.last_seen}`);
  }
}
```

---

## Эндпоинты статической информации

### 1. Получить информацию о сервере

Возвращает статическую информацию о сервере (hostname, OS, hardware).

**Эндпоинт:**
```
GET /api/servers/{server_id}/static-info
GET /api/servers/{server_id}/static-info/server
GET /api/servers/{server_id}/static-info/hardware
GET /api/servers/{server_id}/static-info/network
GET /api/servers/{server_id}/static-info/disks
```

**Ответ (полная информация):**
```json
{
  "server_id": "srv_38822172",
  "hostname": "web-server-01",
  "os": "Ubuntu 25.10",
  "kernel": "6.11.0",
  "architecture": "x86_64",
  "hardware": {
    "cpu_model": "AMD Ryzen 7 7700X",
    "cpu_cores": 8,
    "cpu_threads": 16,
    "cpu_base_frequency_mhz": 4500,
    "memory_total_gb": 29.97,
    "motherboard": "ASUS ROG STRIX B650-A",
    "gpu": "NVIDIA GeForce RTX 4070"
  },
  "disks": [
    {
      "device": "/dev/nvme0n1",
      "model": "Samsung 980 PRO",
      "size_gb": 500,
      "type": "NVMe SSD"
    }
  ],
  "network_interfaces": [
    {
      "name": "enp111s0",
      "mac": "00:11:22:33:44:55",
      "speed_mbps": 1000,
      "type": "ethernet"
    }
  ],
  "created_at": "2026-03-20T10:00:00Z",
  "updated_at": "2026-03-22T12:00:00Z"
}
```

**Ответ (только сервер):**
```
GET /api/servers/{server_id}/static-info/server
```
```json
{
  "hostname": "web-server-01",
  "os": "Ubuntu 25.10",
  "kernel": "6.11.0",
  "architecture": "x86_64"
}
```

**Ответ (только hardware):**
```
GET /api/servers/{server_id}/static-info/hardware
```
```json
{
  "cpu_model": "AMD Ryzen 7 7700X",
  "cpu_cores": 8,
  "cpu_threads": 16,
  "cpu_base_frequency_mhz": 4500,
  "memory_total_gb": 29.97
}
```

---

## Эндпоинты истории метрик

### 1. Получить историю метрик с автоматической гранулярностью

Возвращает историю метрик с автоматическим выбором гранулярности в зависимости от временного диапазона.

**Эндпоинт:**
```
GET /api/servers/{server_id}/metrics/tiered
```

**Параметры запроса:**
- `start` (string, required) - Начало периода в формате RFC3339
- `end` (string, required) - Конец периода в формате RFC3339

**Автоматическая гранулярность:**
- < 1 час → 1 минута
- 1-6 часов → 5 минут
- 6-24 часа → 10 минут
- > 24 часа → 1 час

**Ответ:**
```json
{
  "server_id": "srv_38822172",
  "granularity": "5m",
  "start_time": "2026-03-22T10:00:00Z",
  "end_time": "2026-03-22T14:00:00Z",
  "total_points": 48,
  "data_points": [
    {
      "time": "2026-03-22T10:00:00Z",
      "cpu_avg": 25.5,
      "cpu_max": 45.2,
      "cpu_min": 10.1,
      "memory_avg": 65.2,
      "memory_max": 70.5,
      "memory_min": 60.1,
      "disk_avg": 45.1,
      "network_avg": 3.5,
      "temperature_avg": 48.5,
      "temperature_max": 55.0
    }
  ]
}
```

**Пример запроса:**
```javascript
const start = new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(); // 6 часов назад
const end = new Date().toISOString();

const response = await fetch(
  `/api/servers/srv_38822172/metrics/tiered?start=${start}&end=${end}`
);
const history = await response.json();

console.log(`Granularity: ${history.granularity}`);
console.log(`Total points: ${history.total_points}`);
```

---

### 2. Получить метрики реального времени

Возвращает метрики за последние 15 минут с гранулярностью 1 минута.

**Эндпоинт:**
```
GET /api/servers/{server_id}/metrics/realtime
```

**Ответ:**
```json
{
  "server_id": "srv_38822172",
  "granularity": "1m",
  "time_range": "15m",
  "data_points": [
    {
      "time": "2026-03-22T12:06:00Z",
      "cpu_avg": 25.5,
      "memory_avg": 65.2
    }
  ]
}
```

---

### 3. Получить метрики для дашборда

Возвращает метрики за последние 24 часа с гранулярностью 10 минут.

**Эндпоинт:**
```
GET /api/servers/{server_id}/metrics/dashboard
```

**Ответ:**
```json
{
  "server_id": "srv_38822172",
  "granularity": "10m",
  "time_range": "24h",
  "total_points": 144,
  "data_points": [...]
}
```

---

### 4. Сравнить метрики нескольких серверов

Возвращает метрики для сравнения нескольких серверов.

**Эндпоинт:**
```
GET /api/servers/{server_id}/metrics/comparison
```

**Параметры запроса:**
- `compare_with` (string) - ID серверов для сравнения (через запятую)
- `start` (string) - Начало периода
- `end` (string) - Конец периода

**Пример:**
```
GET /api/servers/srv_111/metrics/comparison?compare_with=srv_222,srv_333&start=2026-03-22T10:00:00Z&end=2026-03-22T14:00:00Z
```

---

## Эндпоинты алертов

### 1. Получить активные алерты

Возвращает список активных алертов для сервера.

**Эндпоинт:**
```
GET /api/servers/{server_id}/alerts
```

**Ответ:**
```json
{
  "server_id": "srv_38822172",
  "active_alerts": [
    {
      "id": "alert_123",
      "type": "cpu_high",
      "severity": "warning",
      "message": "CPU usage above 80%",
      "value": 85.5,
      "threshold": 80,
      "triggered_at": "2026-03-22T12:15:00Z",
      "duration_seconds": 300
    },
    {
      "id": "alert_124",
      "type": "disk_space",
      "severity": "critical",
      "message": "Disk space above 90%",
      "value": 92.1,
      "threshold": 90,
      "triggered_at": "2026-03-22T11:00:00Z",
      "duration_seconds": 4500
    }
  ],
  "total_active": 2
}
```

**Типы алертов:**
- `cpu_high` - Высокая загрузка CPU
- `memory_high` - Высокое использование памяти
- `disk_space` - Заканчивается место на диске
- `temperature_high` - Высокая температура
- `server_offline` - Сервер недоступен

**Уровни severity:**
- `info` - Информационный
- `warning` - Предупреждение
- `critical` - Критический

---

### 2. Создать правило алерта

Создает новое правило для алертов.

**Эндпоинт:**
```
POST /api/servers/{server_id}/alerts/rules
```

**Тело запроса:**
```json
{
  "type": "cpu_high",
  "threshold": 80,
  "duration_seconds": 300,
  "severity": "warning",
  "enabled": true,
  "notification_channels": ["telegram", "email"]
}
```

**Ответ:**
```json
{
  "rule_id": "rule_456",
  "created_at": "2026-03-22T12:30:00Z",
  "status": "active"
}
```

---

## Эндпоинты управления серверами

### 1. Получить список серверов

Возвращает список всех серверов пользователя.

**Эндпоинт:**
```
GET /api/servers
```

**Параметры запроса:**
- `status` (string, optional) - Фильтр по статусу: `online`, `offline`, `all`
- `limit` (int, optional) - Количество результатов (по умолчанию 50)
- `offset` (int, optional) - Смещение для пагинации

**Ответ:**
```json
{
  "servers": [
    {
      "server_id": "srv_111",
      "hostname": "web-server-01",
      "status": "online",
      "last_seen": "2026-03-22T12:30:00Z",
      "cpu_percent": 25.5,
      "memory_percent": 65.2
    },
    {
      "server_id": "srv_222",
      "hostname": "db-server-01",
      "status": "offline",
      "last_seen": "2026-03-22T10:00:00Z",
      "cpu_percent": 0,
      "memory_percent": 0
    }
  ],
  "total": 2,
  "limit": 50,
  "offset": 0
}
```

---

### 2. Управление источниками данных

Добавление/удаление источников уведомлений (Telegram, Discord и т.д.).

**Добавить источник:**
```
POST /api/servers/{server_id}/sources
```

**Тело запроса:**
```json
{
  "source_type": "telegram",
  "identifier": "123456789",
  "metadata": {
    "username": "@johndoe",
    "chat_id": "123456789"
  }
}
```

**Получить источники:**
```
GET /api/servers/{server_id}/sources
```

**Ответ:**
```json
{
  "server_id": "srv_111",
  "sources": [
    {
      "source_type": "telegram",
      "identifier": "123456789",
      "added_at": "2026-03-20T10:00:00Z",
      "metadata": {
        "username": "@johndoe"
      }
    }
  ]
}
```

**Удалить источник:**
```
DELETE /api/servers/{server_id}/sources/{source_type}
```

---

## Коды ошибок

### HTTP коды ответов:

- `200 OK` - Успешный запрос
- `201 Created` - Ресурс создан
- `400 Bad Request` - Неверные параметры запроса
- `401 Unauthorized` - Отсутствует или неверная аутентификация
- `403 Forbidden` - Недостаточно прав доступа
- `404 Not Found` - Ресурс не найден
- `429 Too Many Requests` - Превышен лимит запросов
- `500 Internal Server Error` - Внутренняя ошибка сервера

### Формат ошибки:

```json
{
  "error": "Server not found",
  "code": "SERVER_NOT_FOUND",
  "details": {
    "server_id": "srv_invalid"
  }
}
```

---

## Примеры использования

### React компонент для отображения метрик

```jsx
import React, { useState, useEffect } from 'react';

function ServerMetrics({ serverKey }) {
  const [metrics, setMetrics] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const response = await fetch(
          `/api/servers/by-key/${serverKey}/metrics`,
          {
            headers: {
              'Authorization': `Bearer ${process.env.REACT_APP_API_KEY}`
            }
          }
        );
        const data = await response.json();
        setMetrics(data.metrics);
      } catch (error) {
        console.error('Failed to fetch metrics:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 60000); // Обновление каждую минуту

    return () => clearInterval(interval);
  }, [serverKey]);

  if (loading) return <div>Loading...</div>;
  if (!metrics) return <div>No data</div>;

  return (
    <div className="metrics-card">
      <h3>Server Metrics</h3>
      <div className="metric">
        <span>CPU:</span>
        <span>{metrics.cpu_percent.toFixed(1)}%</span>
      </div>
      <div className="metric">
        <span>Memory:</span>
        <span>{metrics.memory_percent.toFixed(1)}%</span>
      </div>
      <div className="metric">
        <span>Disk:</span>
        <span>{metrics.disk_percent.toFixed(1)}%</span>
      </div>
      <div className="metric">
        <span>Temperature:</span>
        <span>{metrics.temperature_celsius}°C</span>
      </div>
    </div>
  );
}

export default ServerMetrics;
```

---

### Vue.js компонент для истории метрик

```vue
<template>
  <div class="metrics-chart">
    <h3>CPU History (Last 6 hours)</h3>
    <canvas ref="chartCanvas"></canvas>
  </div>
</template>

<script>
import { Chart } from 'chart.js';

export default {
  name: 'MetricsChart',
  props: ['serverId'],
  data() {
    return {
      chart: null
    };
  },
  async mounted() {
    await this.loadMetrics();
  },
  methods: {
    async loadMetrics() {
      const end = new Date().toISOString();
      const start = new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString();
      
      const response = await fetch(
        `/api/servers/${this.serverId}/metrics/tiered?start=${start}&end=${end}`,
        {
          headers: {
            'Authorization': `Bearer ${process.env.VUE_APP_API_KEY}`
          }
        }
      );
      
      const data = await response.json();
      this.renderChart(data.data_points);
    },
    
    renderChart(dataPoints) {
      const ctx = this.$refs.chartCanvas.getContext('2d');
      
      this.chart = new Chart(ctx, {
        type: 'line',
        data: {
          labels: dataPoints.map(p => new Date(p.time).toLocaleTimeString()),
          datasets: [{
            label: 'CPU Average',
            data: dataPoints.map(p => p.cpu_avg),
            borderColor: 'rgb(75, 192, 192)',
            tension: 0.1
          }]
        },
        options: {
          responsive: true,
          scales: {
            y: {
              beginAtZero: true,
              max: 100
            }
          }
        }
      });
    }
  }
};
</script>
```

---

### Vanilla JavaScript для статуса сервера

```javascript
class ServerStatusMonitor {
  constructor(serverKey, apiKey) {
    this.serverKey = serverKey;
    this.apiKey = apiKey;
    this.statusElement = document.getElementById('server-status');
  }

  async checkStatus() {
    try {
      const response = await fetch(
        `/api/servers/by-key/${this.serverKey}/status`,
        {
          headers: {
            'Authorization': `Bearer ${this.apiKey}`
          }
        }
      );

      const status = await response.json();
      this.updateUI(status);
    } catch (error) {
      console.error('Failed to check status:', error);
      this.showError();
    }
  }

  updateUI(status) {
    const statusClass = status.online ? 'online' : 'offline';
    const statusText = status.online ? 'Online' : 'Offline';
    
    this.statusElement.innerHTML = `
      <div class="status-indicator ${statusClass}"></div>
      <span>${statusText}</span>
      <small>Last seen: ${new Date(status.last_seen).toLocaleString()}</small>
      <small>Agent: v${status.agent_version}</small>
    `;
  }

  showError() {
    this.statusElement.innerHTML = `
      <div class="status-indicator error"></div>
      <span>Error checking status</span>
    `;
  }

  startMonitoring(intervalSeconds = 30) {
    this.checkStatus();
    setInterval(() => this.checkStatus(), intervalSeconds * 1000);
  }
}

// Использование
const monitor = new ServerStatusMonitor('key_f1a5af38', 'your_api_key');
monitor.startMonitoring(30); // Проверка каждые 30 секунд
```

---

## Rate Limiting

API имеет следующие лимиты запросов:

- **Бесплатный план:** 100 запросов/минуту
- **Pro план:** 1000 запросов/минуту
- **Enterprise план:** Без ограничений

При превышении лимита API вернет:

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

**Заголовки ответа:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1774182140
```

---

## WebSocket API (Real-time)

Для получения метрик в реальном времени используйте WebSocket соединение:

**Подключение:**
```javascript
const ws = new WebSocket('wss://api.servereye.com/ws');

ws.onopen = () => {
  // Аутентификация
  ws.send(JSON.stringify({
    type: 'auth',
    api_key: 'your_api_key'
  }));
  
  // Подписка на сервер
  ws.send(JSON.stringify({
    type: 'subscribe',
    server_id: 'srv_38822172'
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  if (data.type === 'metrics') {
    console.log('New metrics:', data.metrics);
    updateDashboard(data.metrics);
  }
};
```

---

## Поддержка

- **Email:** support@servereye.com
- **Документация:** https://docs.servereye.com
- **GitHub:** https://github.com/servereye/api
- **Discord:** https://discord.gg/servereye

---

**Последнее обновление:** 22 марта 2026  
**Версия документации:** 1.0.0
