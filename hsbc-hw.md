## Instructions to the candidate for the take home task

Write a simple (and possibly not very secure) authentication and authorization service. The service allows users to be authenticated, and authorizes different
behavior. No real knowledge of cryptography is assumed. You should complete this by yourself and you are free to use whatever resources/reference materials you see fit. The completed work should be returned via email or public Github no more than 48 hours after receiving this assignment. The exercise itself is not meant to take more than 1-2 hours to complete. Once the submission has been reviewed, a follow-up discussion will be arranged.

### Directions

* Write in Java
* Don't use any existing security classes from Java (other than hashing/encryption if you need to)
* The main entities are:
  * Users
  * Roles - special entities, that can be associated with individual users. Each user can have multiple roles assigned to it
* Keep all data in memory, no persistence storage is required
* No need to sign tokens in any special way, it is assumed that the communication channel between the API and the consumer is secure
* The main points to address are:
  * Clean API
  * Performance of all the main operations
  * Thorough testing (yes, that includes token expiry)
* The deliverable should be a self-contained project we can easily open and run the tests in IntelliJ
* Use Maven or Gradle
* Remember, the auth token is something that can be passed around outside your service. While you are welcome to implement it internally the way you like, the value you pass around between calls should be some primitive type, long or string
* If you use external libraries outside of the standard JDK, please mention in the README and explain their purpose

### API to implement

Feel free to name your functions as you see fit, as long as the action is clearly stated.

* Create user
  * User name
  * Password-to be stored in some encrypted form
  * Should fail if the user already exists
* Delete user
  * User
  * password?????
  * Should fail if the user doesn't exist
* Create role
  * Role name
  * Should fail if the role already exists
* Delete role
 * Role
 * Should fail if the role doesn't exist
* Add role to user
 * User
 * Role
 * If the role is already associated with the user, nothing should happen
* Authenticate
  * user name
  * password
  * return a special "secret" auth token or error, if not found. The token is only valid for pre-configured time(2h)
* Invalidate
  * auth token
  * returns nothing, the token is no longer valid after the call. Handles correctly the case of invalid token given as input
* Check role
  * auth token
  * role
  * returns true if the user, identified by the token, belongs to the role, false otherwise; error if token is invalid expired etc
* All roles
  * auth token
  * returns all roles for the user, error if token is invalid
