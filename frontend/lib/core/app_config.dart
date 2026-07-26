class AppConfig {
  const AppConfig._();

  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080/api/v1',
  );

  static Uri get websocketUri {
    final api = Uri.parse(apiBaseUrl);
    return api.replace(
      scheme: api.scheme == 'https' ? 'wss' : 'ws',
      path: '/ws',
      query: null,
    );
  }
}
