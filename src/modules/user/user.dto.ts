import { ServerDto } from '../server/server.dto'

export class UserDto {
  id?: number
  name?: string
  email?: string
  password?: string
  servers?: ServerDto[]
}
