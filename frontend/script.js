document.getElementById('findOrderBtn').addEventListener('click', findOrder);

function formatJSON(obj) {
    return JSON.stringify(obj, null, 2);
}

async function findOrder() {
    const orderId = document.getElementById('orderId').value.trim();
    const resultDiv = document.getElementById('result');
    resultDiv.className = '';
    resultDiv.textContent = 'Загрузка...';

    if (!orderId) {
        resultDiv.textContent = 'Пожалуйста, введите order_uid';
        resultDiv.className = 'error';
        return;
    }

    try {
        const response = await fetch(`http://localhost:8081/order/${orderId}`);
        if (!response.ok) {
            resultDiv.textContent = 'Заказ не найден';
            resultDiv.className = 'error';
            return;
        }
        const data = await response.json();
        resultDiv.textContent = formatJSON(data);
        resultDiv.className = 'success';
    } catch (err) {
        resultDiv.textContent = 'Ошибка запроса: ' + err;
        resultDiv.className = 'error';
    }
} 