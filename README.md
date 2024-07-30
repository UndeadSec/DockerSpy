# DockerSpy
DockerSpy searches for images on Docker Hub and extracts sensitive information such as authentication secrets, private keys, and more.

<p align="center">
<img src="https://github.com/UndeadSec/DockerSpy/blob/main/screenshot/sc.png?raw=true"/>
</p>

### What is Docker?

Docker is an open-source platform that automates the deployment, scaling, and management of applications using containerization technology. Containers allow developers to package an application and its dependencies into a single, portable unit that can run consistently across various computing environments. Docker simplifies the development and deployment process by ensuring that applications run the same way regardless of where they are deployed.

### About Docker Hub

Docker Hub is a cloud-based repository where developers can store, share, and distribute container images. It serves as the largest library of container images, providing access to both official images created by Docker and community-contributed images. Docker Hub enables developers to easily find, download, and deploy pre-built images, facilitating rapid application development and deployment.

### Why OSINT on Docker Hub?

Open Source Intelligence (OSINT) on Docker Hub involves using publicly available information to gather insights and data from container images and repositories hosted on Docker Hub. This is particularly important for identifying exposed secrets for several reasons:

1. **Security Audits**: By analyzing Docker images, organizations can uncover exposed secrets such as API keys, authentication tokens, and private keys that might have been inadvertently included. This helps in mitigating potential security risks.

2. **Incident Prevention**: Proactively searching for exposed secrets in Docker images can prevent security breaches before they happen, protecting sensitive information and maintaining the integrity of applications.

3. **Compliance**: Ensuring that container images do not expose secrets is crucial for meeting regulatory and organizational security standards. OSINT helps verify that no sensitive information is unintentionally disclosed.

4. **Vulnerability Assessment**: Identifying exposed secrets as part of regular security assessments allows organizations to address these vulnerabilities promptly, reducing the risk of exploitation by malicious actors.

5. **Enhanced Security Posture**: Continuously monitoring Docker Hub for exposed secrets strengthens an organization's overall security posture, making it more resilient against potential threats.

Utilizing OSINT on Docker Hub to find exposed secrets enables organizations to enhance their security measures, prevent data breaches, and ensure the confidentiality of sensitive information within their containerized applications.

- [Thousands of images on Docker Hub leak auth secrets, private keys](https://www.bleepingcomputer.com/news/security/thousands-of-images-on-docker-hub-leak-auth-secrets-private-keys/)
- [Docker Hub images found to expose secrets and private keys](https://www.threatdown.com/blog/docker-hub-images-found-to-expose-secrets-and-private-keys/)

## How DockerSpy Works

DockerSpy obtains information from Docker Hub and uses regular expressions to inspect the content for sensitive information, such as secrets.

## Getting Started

To use DockerSpy, follow these steps:

1. **Installation:** Clone the DockerSpy repository and install the required dependencies.

```bash
git clone https://github.com/UndeadSec/DockerSpy.git && cd DockerSpy && make
```

2. **Usage:** Run DockerSpy from terminal.

```bash
dockerspy
```

## Custom Configurations

To customize DockerSpy configurations, edit the following files:
- [Regular Expressions](src/configs/regex_patterns.json)
- [Ignored File Extensions](src/configs/ignore_extensions.json)

## Disclaimer

DockerSpy is intended for educational and research purposes only. Users are responsible for ensuring that their use of this tool complies with applicable laws and regulations.

## Contribution

Contributions to DockerSpy are welcome! Feel free to submit issues, feature requests, or pull requests to help improve this tool.

## About the Author

DockerSpy is developed and maintained by *Alisson Moretto* (UndeadSec)

I'm a passionate cyber threat intelligence pro who loves sharing insights and crafting cybersecurity tools.

Consider following me:

[![X](https://img.shields.io/badge/X-%23000000.svg?style=for-the-badge&logo=X&logoColor=white)](https://twitter.com/UndeadSec)
[![LinkedIn](https://img.shields.io/badge/linkedin-%230077B5.svg?style=for-the-badge&logo=linkedin&logoColor=white)](https://linkedin.com/in/alissonmoretto)
[![GitHub](https://img.shields.io/badge/github-%23121011.svg?style=for-the-badge&logo=github&logoColor=white)](https://github.com/UndeadSec)

## TODO

### Regular Expressions Enhancement

- [ ] Review and improve existing regular expressions.
- [ ] Ensure that regular expressions adhere to best practices.
- [ ] Check for any potential optimizations in the regex patterns.
- [ ] Test regular expressions with various input scenarios for accuracy.
- [ ] Document any complex or non-trivial regex patterns for better understanding.

## License

DockerSpy is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

### Thanks

Special thanks to [@akaclandestine](https://x.com/akaclandestine) 
