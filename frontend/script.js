async function findOrder() {
    const orderId = document.getElementById('orderId').value.trim();
    const resultDiv = document.getElementById('result');
    resultDiv.textContent = 'Загрузка...';

    if (!orderId) {
        resultDiv.textContent = 'Пожалуйста, введите order_uid';
        return;
    }

    try {
        const response = await fetch(`http://localhost:8081/order/${orderId}`);
        if (!response.ok) {
            resultDiv.textContent = 'Заказ не найден';
            return;
        }
        const data = await response.json();
        resultDiv.textContent = JSON.stringify(data, null, 2);
    } catch (err) {
        resultDiv.textContent = 'Ошибка запроса: ' + err;
    }
}