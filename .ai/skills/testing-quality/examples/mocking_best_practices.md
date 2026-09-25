# Example: Mocking Best Practices

When writing unit tests, you should isolate the system under test (SUT) from external dependencies.

## The SUT (System Under Test)

```python
import requests

class WeatherService:
    def __init__(self, api_key: str):
        self.api_key = api_key

    def get_temperature(self, city: str) -> float:
        response = requests.get(f"https://api.weather.com/v1/{city}?key={self.api_key}")
        response.raise_for_status()
        data = response.json()
        return data['temp']
```

## The Unit Test

```python
import unittest
from unittest.mock import patch, Mock
from my_app.weather import WeatherService

class TestWeatherService(unittest.TestCase):
    
    @patch('my_app.weather.requests.get')
    def test_get_temperature_returns_correct_value(self, mock_get):
        # Arrange
        mock_response = Mock()
        mock_response.json.return_value = {'temp': 22.5}
        mock_response.raise_for_status.return_value = None
        mock_get.return_value = mock_response
        
        service = WeatherService(api_key="fake-key")
        
        # Act
        temp = service.get_temperature("London")
        
        # Assert
        self.assertEqual(temp, 22.5)
        mock_get.assert_called_once_with("https://api.weather.com/v1/London?key=fake-key")
```

**Key Takeaway**: We patch `requests.get` where it is *used* (`my_app.weather.requests.get`), not where it is defined, allowing us to test `WeatherService` without making actual network calls.
