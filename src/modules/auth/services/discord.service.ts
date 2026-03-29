import { BadRequestException, Injectable } from '@nestjs/common'
import { Response } from 'express'
import { UserDto } from '../../user/user.dto'
import { Payload } from '../interfaces/payload.interface'
import { catchError, firstValueFrom } from 'rxjs'
import { UserService } from '../../user/user.service'
import { ConfigService } from '@nestjs/config'
import { HttpService } from '@nestjs/axios'
import { AuthService } from '../auth.service'

@Injectable()
export class DiscordService {
  private readonly DISCORD_CLIENT_ID: string
  private readonly DISCORD_CLIENT_SECRET: string
  private readonly DISCORD_REDIRECT_URI: string

  constructor(
    private readonly userService: UserService,
    private readonly configService: ConfigService,
    private readonly httpService: HttpService,
    private readonly authService: AuthService,
  ) {
    this.DISCORD_CLIENT_ID = this.configService.getOrThrow('DISCORD_CLIENT_ID')
    this.DISCORD_CLIENT_SECRET = this.configService.getOrThrow(
      'DISCORD_CLIENT_SECRET'
    )
    this.DISCORD_REDIRECT_URI = this.configService.getOrThrow(
      'DISCORD_REDIRECT_URI'
    )
  }

  async discordCallback(res: Response, code: string) {
    const response = await this.getDiscordResponse(code)

    const accessToken = response.data.access_token as string
    const refreshToken = response.data.refresh_token as string

    const userResponse = await this.getDiscordUser(accessToken)

    const userResponseId = userResponse.data.user.id as string
    const userResponseUsername = userResponse.data.user.username as string
    const userResponseAvatar = userResponse.data.user.avatar as string

    const existingUser: UserDto | null = await this.userService.findOneOrNull({
      discordId: userResponseId
    })

    const expiresIn = response.data.expires_in as number

    this.authService.setCookie(res, refreshToken, new Date(Date.now() + expiresIn))

    if (existingUser) {
      const payload: Payload = { id: existingUser.id! }
      return this.authService.auth(res, payload)
    }

    const avatarUrl = `https://cdn.discordapp.com/avatars/${userResponseId}/${userResponseAvatar}.png`

    const user: UserDto = await this.userService.discordRegister({
      name: userResponseUsername,
      discordId: userResponseId,
      avatar: avatarUrl
    })

    const payload: Payload = { id: user.id! }

    return this.authService.auth(res, payload)
  }

  private getDiscordResponse(code: string) {
    return firstValueFrom(
      this.httpService
        .post(
          '/token',
          {
            grant_type: 'authorization_code',
            code,
            redirect_uri: this.DISCORD_REDIRECT_URI
          },
          {
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded'
            },
            auth: {
              username: this.DISCORD_CLIENT_ID,
              password: this.DISCORD_CLIENT_SECRET
            }
          }
        )
        .pipe(
          catchError(() => {
            throw new BadRequestException('Discord OAuth failed')
          })
        )
    )
  }

  private getDiscordUser(accessToken: string) {
    return firstValueFrom(
      this.httpService
        .get('/@me', {
          headers: {
            Authorization: `Bearer ${accessToken}`
          }
        })
        .pipe(
          catchError(() => {
            throw new BadRequestException('Discord user getting failed')
          })
        )
    )
  }
}
