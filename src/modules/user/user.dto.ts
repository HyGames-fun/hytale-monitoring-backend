import { ServerDto } from '../server/server.dto'

export class UserDto {
  id?: number
  name?: string
  email?: string | null
  avatar?: string | null
  password?: string | null
  discordId?: string | null
  servers?: ServerDto[]
}

export class UpdateUserDto {
  name?: string
  email?: string | null
  avatar?: string | null
  password?: string | null
  discordId?: string | null
}
