# Planifest Infrastructure Standards

> Infrastructure is code. It is versioned, reviewed, tested, and deployed through the same pipeline as application code. These standards ensure agent-generated IaC is production-grade.

---

## 1. Infrastructure as Code

- All infrastructure must be defined in code - no manual console changes
- Use the IaC tool declared in the stack (Terraform, Pulumi, CDK)
- IaC files live at `src/{component-id}/infra/` or a dedicated infrastructure component
- All configuration is parameterized - no hardcoded values for environment, region, or credentials

---

## 2. Environment Separation

| Environment | Purpose | Data | Access |
|-------------|---------|------|--------|
| Development | Local development and testing | Synthetic/seed data | Developer |
| Staging | Pre-production validation | Anonymized production-like data | Team |
| Production | Live traffic | Real data | Restricted |

- Environment-specific values use variables/parameters, never conditional logic in IaC
- Production infrastructure changes require human approval

---

## 3. Security Defaults

- **Least privilege:** IAM roles grant the minimum permissions required
- **No public access by default:** Resources are private unless explicitly declared public
- **Encryption at rest:** Enabled for all data stores
- **Encryption in transit:** TLS required for all network communication
- **No default credentials:** All passwords, keys, and tokens are generated and stored in a secrets manager
- **Audit logging:** CloudTrail/Cloud Audit Logs enabled for all services

---

## 4. Networking

- Use VPC/VNet with private subnets for compute and data
- Public subnets only for load balancers and NAT gateways
- Security groups/firewall rules follow least-privilege - only open required ports
- No `0.0.0.0/0` ingress rules except for public load balancers on ports 80/443

---

## 5. Compute

- Use the compute model declared in the stack (Cloud Run, Lambda, ECS, K8s)
- Configure health checks for all services
- Set resource limits (CPU, memory) to prevent runaway costs
- Configure auto-scaling based on metrics identified in the SLO definitions

---

## 6. Data Stores

- Automated backups with retention matching the operational model
- Point-in-time recovery enabled for relational databases
- Connection pooling configured for all database connections
- Read replicas for read-heavy workloads (if the SLO requires it)

---

*Referenced by codegen-agent and security-agent. Source of truth: `planifest-framework/standards/infrastructure-standards.md`*
