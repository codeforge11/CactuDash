function sendPasswd() {
    document.addEventListener('DOMContentLoaded', function () {
        const errorLabel = document.getElementById('errorMessage');
        const form = document.getElementById('auth-form');
        const passwd1 = document.getElementById('passwd1');
        const passwd2 = document.getElementById('passwd2');

        function showError(message) {
            errorLabel.textContent = message;
            errorLabel.style.display = 'block';
            setTimeout(() => {
                errorLabel.style.display = 'none';
            }, 3000);
        }

        form.addEventListener('submit', function (e) {
            if (passwd1.value.length < 8) {
                e.preventDefault();
                showError('Password must be at least 8 characters!');
            } else if (passwd1.value !== passwd2.value) {
                e.preventDefault();
                showError('Passwords must be same!');
            }
        });

        [passwd1, passwd2].forEach(input => input.addEventListener('input', () => {
            errorLabel.style.display = 'none';
        })
        );
    });
}

sendPasswd()