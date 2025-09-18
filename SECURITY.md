# Security Policy

## Supported Versions

We actively maintain and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| 1.x.x   | :x:                |

## Security Features

This project implements multiple security measures:

### Input Validation & Sanitization
- All user inputs are validated using regex patterns
- SQL injection prevention through parameterized queries
- Log injection prevention with structured logging
- Path traversal protection with allowlisting

### Network Security
- SSRF protection with URL allowlisting for external APIs
- HTTPS-only external communications
- Secure HTTP client configuration with timeouts

### Data Protection
- Environment-based configuration for sensitive data
- No hardcoded credentials in source code
- Secure database connection management
- Proper resource cleanup and connection pooling

### Application Security
- Concurrency control to prevent resource exhaustion
- Graceful error handling without information disclosure
- Structured logging without sensitive data exposure
- Dependency vulnerability scanning

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do Not Create Public Issues
Please do not create public GitHub issues for security vulnerabilities.

### 2. Contact Us Privately
Send an email to: **security@yulia-lingo.com** (or create a private issue if email is not available)

Include the following information:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact assessment
- Any suggested fixes (if available)

### 3. Response Timeline
- **Initial Response**: Within 48 hours
- **Vulnerability Assessment**: Within 7 days
- **Fix Development**: Within 30 days (depending on severity)
- **Public Disclosure**: After fix is deployed and users have time to update

### 4. Severity Levels

We classify vulnerabilities using the following severity levels:

#### Critical
- Remote code execution
- SQL injection leading to data breach
- Authentication bypass

#### High
- Cross-site scripting (XSS)
- Privilege escalation
- Sensitive data exposure

#### Medium
- Information disclosure
- Denial of service
- CSRF vulnerabilities

#### Low
- Minor information leaks
- Non-exploitable security misconfigurations

## Security Best Practices for Users

### Environment Configuration
1. Use strong, unique passwords for database connections
2. Regularly rotate API keys and tokens
3. Use environment variables for all sensitive configuration
4. Enable database SSL/TLS in production

### Deployment Security
1. Run the application with minimal privileges
2. Use container security scanning
3. Keep dependencies updated
4. Monitor logs for suspicious activity
5. Implement network segmentation

### Monitoring
1. Enable structured logging
2. Monitor for unusual patterns in bot interactions
3. Set up alerts for database connection failures
4. Track resource usage patterns

## Security Updates

Security updates will be:
1. Released as patch versions (e.g., 2.1.1 → 2.1.2)
2. Documented in release notes with severity information
3. Announced through GitHub releases and security advisories

## Acknowledgments

We appreciate the security research community and will acknowledge researchers who responsibly disclose vulnerabilities (with their permission).

## Contact

For security-related questions or concerns:
- Email: security@yulia-lingo.com
- GitHub: Create a private security advisory
- Telegram: @zufarexplained (for urgent issues only)

---

**Note**: This security policy is subject to updates. Please check back regularly for the latest information.