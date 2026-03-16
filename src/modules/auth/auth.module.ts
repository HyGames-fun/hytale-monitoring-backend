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
  providers: [AuthService, JwtStrategy]
})
export class AuthModule {}
