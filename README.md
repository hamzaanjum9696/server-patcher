# server-patching

## Goal:
1. Prepare server for Patching
2. After Patching, restart all the required services

## Preparation Phase
- There can be multiple types of servers e.g. WEB, Apache, BE
- Check the type of server to act accordingly:
    - **Apache Servers**
        - Running apache Snapshot
        - Send it via email for record and save in safe directory e.g. /u/Server_Patching_Automation/
        - Stop all the apache servers
        - Notify via email that the server is ready for patching
        - After server is patched, start these apaches
        - Verify the number of running apache
        - Check logs for hits to verify if apache is handling requests now