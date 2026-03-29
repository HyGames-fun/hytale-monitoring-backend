import { Module } from '@nestjs/common'
import { AuthService } from './auth.service'
import { AuthController } from './auth.controller'
import { JwtModule } from '@nestjs/jwt'
import { ConfigService } from '@nestjs/config'
import { getJwtConfig } from '../../../configs/jwt.config'
import { UserModule } from '../user/user.module'
import { JwtStrategy } from './strategies/jwt.strategy'
import { HttpModule } from '@nestjs/axios'
import { getHttpConfig } from '../../../configs/http.config'
import { TokenService } from './services/token.service'
import { VerificationService } from './services/verification.service'
import { MailService } from './services/mail.service'
import { DiscordService } from './services/discord.service'

@Module({
  imports: [
    JwtModule.registerAsync({
      useFactory: getJwtConfig,
      inject: [ConfigService]
    }),
    HttpModule.registerAsync({
      useFactory: getHttpConfig,
      inject: [ConfigService]
    }),
    UserModule
  ],
  controllers: [AuthController],
  providers: [
    AuthService,
    JwtStrategy,
    TokenService,
    VerificationService,
    MailService,
    DiscordService
  ]
})
export class AuthModule {}
