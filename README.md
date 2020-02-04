## MYRA-DYN
Myra protected DYNDNS

### CONFIGURATION:
Replace the apikey and the secret defined in the config.yml file
config.yml
- apikey: 'ADD_YOUR_API_KEY'
- secret: 'ADD_YOUR_SECRET'
        
### BUILDING:
running 'make' will create a binary in ./bin/ called myra-dyn
        
### USAGE:
Add the domains(!) you want to update using your machines IP address.
```
./bin/myra-dyn your-domain.tld
./bin/myra-dyn example1.tld example2.tld example3.tld
```

### NOTE:
- Right now, myra-dyn only works for whole domains. It is not possible to specify a single subdomain for dynamic DNS usage.
- If your machines IP changes from IPv4 to IPv6 (or the other way around), the DNS recordType will be changed, to (e.g. A instead of AAAA)